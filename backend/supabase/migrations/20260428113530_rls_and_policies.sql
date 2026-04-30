-- RLS
alter table agencies enable row level security;
alter table users     enable row level security;
alter table tasks     enable row level security;

-- Helper: get the agency_id of the currently authenticated active user
create function auth_user_agency_id()
returns uuid language sql stable security definer
set search_path = ''
as $$
  select agency_id from public.users
  where id = auth.uid() and deleted_at is null
$$;

-- Helper: true if the current user has role = 'admin' in their JWT app_metadata
create function auth_user_is_admin()
returns boolean language sql stable security definer
set search_path = ''
as $$
  select coalesce(
    (auth.jwt() -> 'app_metadata' ->> 'role') = 'admin',
    false
  )
$$;

-- Trigger: block non-admins from changing assigned_to (RLS cannot compare OLD vs NEW)
create function prevent_assignee_change_by_non_admin()
returns trigger language plpgsql security definer
set search_path = ''
as $$
begin
  if old.assigned_to is distinct from new.assigned_to
     and not public.auth_user_is_admin()
  then
    raise exception 'only admins can change the task assignee';
  end if;
  return new;
end;
$$;

create trigger tasks_guard_assignee
  before update on public.tasks
  for each row execute function prevent_assignee_change_by_non_admin();

-- agencies: members can read their own non-deleted agency
create policy "agencies: members can read own agency"
  on agencies for select
  using (id = auth_user_agency_id() and deleted_at is null);

-- users: members can read active users in the same agency
create policy "users: read same agency"
  on users for select
  using (agency_id = auth_user_agency_id() and deleted_at is null);

-- users: only admins can create user profile rows
create policy "users: admins can insert"
  on users for insert
  with check (
    auth_user_is_admin()
    and agency_id = auth_user_agency_id()
  );

-- users: a user can update their own active profile
create policy "users: update own profile"
  on users for update
  using (id = auth.uid() and deleted_at is null)
  with check (id = auth.uid());

-- users: only admins can hard-delete user rows
create policy "users: admins can delete"
  on users for delete
  using (
    auth_user_is_admin()
    and agency_id = auth_user_agency_id()
  );

-- tasks: members can read tasks in their active agency
create policy "tasks: read own agency"
  on tasks for select
  using (agency_id = auth_user_agency_id());

-- tasks: members can create tasks in their own active agency
create policy "tasks: insert in own agency"
  on tasks for insert
  with check (
    agency_id = auth_user_agency_id()
    and created_by = auth.uid()
  );

-- tasks: members can update tasks in their active agency
create policy "tasks: update in own agency"
  on tasks for update
  using  (agency_id = auth_user_agency_id())
  with check (agency_id = auth_user_agency_id());

-- tasks: only admins of the agency can delete tasks
create policy "tasks: admins can delete"
  on tasks for delete
  using (
    auth_user_is_admin()
    and agency_id = auth_user_agency_id()
  );
