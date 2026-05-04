-- Enable full replica identity on tasks so UPDATE and DELETE events include
-- all columns (including agency_id) in the replication stream. Without this,
-- the realtime filter `agency_id=eq.{id}` only works for INSERT events.
alter table public.tasks replica identity full;

-- Add tasks to the realtime publication so Supabase broadcasts changes.
alter publication supabase_realtime add table public.tasks;
