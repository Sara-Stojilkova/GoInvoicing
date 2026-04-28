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

-- Helper: get agency_id directly from the JWT claims (app_metadata)
create function jwt_agency_id()
returns uuid language sql stable
set search_path = ''
as $$
  select (auth.jwt() -> 'app_metadata' ->> 'agency_id')::uuid
$$;

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
    and agency_id = jwt_agency_id()
  )
  with check (
    auth_user_is_admin()
    and agency_id = jwt_agency_id()
  );

-- Helper: true if the current user is active (not deleted, activated)
create function auth_user_is_active()
returns boolean language sql stable security definer
set search_path = ''
as $$
  select exists (
    select 1 from public.users
    where id = auth.uid() and activated = true and deleted_at is null
  )
$$;

-- Drop and recreate all agency-scoped policies to use jwt_agency_id().
-- Policies that permit access also require auth_user_is_active() since
-- jwt_agency_id() reads the JWT directly and does not check activation.

drop policy if exists "agencies: members can read own agency"  on public.agencies;
drop policy if exists "users: read same agency"                on public.users;
drop policy if exists "users: admins can insert"               on public.users;
drop policy if exists "users: admins can delete"               on public.users;
drop policy if exists "tasks: read own agency"                 on public.tasks;
drop policy if exists "tasks: insert in own agency"            on public.tasks;
drop policy if exists "tasks: update in own agency"            on public.tasks;
drop policy if exists "tasks: admins can delete"               on public.tasks;

-- agencies
create policy "agencies: members can read own agency"
  on public.agencies for select
  using (id = jwt_agency_id() and deleted_at is null and auth_user_is_active());

-- users
create policy "users: read same agency"
  on public.users for select
  using (agency_id = jwt_agency_id() and deleted_at is null and auth_user_is_active());

create policy "users: admins can insert"
  on public.users for insert
  with check (auth_user_is_admin() and agency_id = jwt_agency_id());

create policy "users: admins can delete"
  on public.users for delete
  using (auth_user_is_admin() and agency_id = jwt_agency_id());

-- tasks
create policy "tasks: read own agency"
  on public.tasks for select
  using (agency_id = jwt_agency_id() and auth_user_is_active());

create policy "tasks: insert in own agency"
  on public.tasks for insert
  with check (
    agency_id = jwt_agency_id()
    and created_by = auth.uid()
    and auth_user_is_active()
  );

create policy "tasks: update in own agency"
  on public.tasks for update
  using  (agency_id = jwt_agency_id() and auth_user_is_active())
  with check (agency_id = jwt_agency_id() and auth_user_is_active());

create policy "tasks: admins can delete"
  on public.tasks for delete
  using (auth_user_is_admin() and agency_id = jwt_agency_id());
