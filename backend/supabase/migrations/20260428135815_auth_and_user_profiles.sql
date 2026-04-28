-- Add activated column: new users start inactive until an agency admin activates them
alter table public.users add column activated boolean not null default false;

-- Trigger: create a public.users profile row whenever a new auth.users row is inserted.
create function handle_new_user()
returns trigger language plpgsql security definer
set search_path = ''
as $$
begin
  insert into public.users (id, agency_id, full_name)
  values (
    new.id,
    (new.raw_user_meta_data ->> 'agency_id')::uuid,
    new.raw_user_meta_data ->> 'full_name'
  );
  return new;
end;
$$;

create trigger on_auth_user_created
  after insert on auth.users
  for each row execute function handle_new_user();

-- Replace auth_user_agency_id() to also require activated = true.
-- Unactivated users get NULL, which causes all agency-scoped policies to deny access.
create or replace function auth_user_agency_id()
returns uuid language sql stable security definer
set search_path = ''
as $$
  select agency_id from public.users
  where id = auth.uid() and deleted_at is null and activated = true
$$;

-- Replace users: update own profile to block self-activation
drop policy if exists "users: update own profile" on public.users;
create policy "users: update own profile"
  on public.users for update
  using  (id = auth.uid() and deleted_at is null)
  with check (id = auth.uid() and activated = (select activated from public.users where id = auth.uid()));

-- users: admins can activate (and otherwise update) any user in their agency
create policy "users: admins can update"
  on public.users for update
  using (
    auth_user_is_admin()
    and agency_id = auth_user_agency_id()
  )
  with check (
    auth_user_is_admin()
    and agency_id = auth_user_agency_id()
  );
