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
