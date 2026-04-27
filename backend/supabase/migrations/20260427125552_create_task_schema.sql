-- Enums
create type task_status as enum ('todo', 'in_progress', 'done');
create type task_priority as enum ('low', 'medium', 'high');

-- Agencies
create table agencies (
  id         uuid primary key default gen_random_uuid(),
  name       text not null,
  created_at timestamptz not null default now(),
  deleted_at timestamptz
);

-- Users (profile table — one row per auth.users entry)
create table users (
  id         uuid primary key references auth.users(id) on delete cascade,
  agency_id  uuid not null references agencies(id) on delete restrict,
  full_name  text,
  created_at timestamptz not null default now(),
  deleted_at timestamptz
);

-- Tasks
create table tasks (
  id          uuid primary key default gen_random_uuid(),
  agency_id   uuid not null references agencies(id) on delete cascade,
  created_by  uuid not null references users(id) on delete restrict,
  assigned_to uuid references users(id) on delete set null,
  title       text not null,
  description text,
  status      task_status not null default 'todo',
  priority    task_priority not null default 'medium',
  due_date    date,
  created_at  timestamptz not null default now(),
  updated_at  timestamptz not null default now()
);
