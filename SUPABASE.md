# Supabase Configuration

## Project details

| Field        | Value                          |
|--------------|--------------------------------|
| Project name | TaskManagementApplication      |
| Project ref  | `mshugzgwjnyyjdkyjwgn`         |
| Region       | AWS eu-west-2 (London)         |
| Dashboard    | https://supabase.com/dashboard/project/mshugzgwjnyyjdkyjwgn |
| API URL      | `https://mshugzgwjnyyjdkyjwgn.supabase.co` |

## Credentials

Connection details live in `backend/.env`. Copy `backend/.env.example` and fill in the values from the Supabase dashboard (Settings → API).

| Variable                    | Where to find it               | Notes                                      |
|-----------------------------|--------------------------------|--------------------------------------------|
| `SUPABASE_URL`              | Settings → API → Project URL   | Safe to commit                             |
| `SUPABASE_ANON_KEY`         | Settings → API → anon/public   | Safe to expose to the browser              |
| `SUPABASE_SERVICE_ROLE_KEY` | Settings → API → service_role  | **Keep secret — never commit or expose**   |
| `DATABASE_URL`              | Settings → Database → Connection string (URI mode) | Used by the Go backend for direct Postgres access |

For local development the Supabase CLI uses its own local keys printed by `supabase start`. Do not mix local keys with production keys.

## Connection strings

**Pooler (recommended for the Go backend):**
```
postgresql://postgres.mshugzgwjnyyjdkyjwgn@aws-1-eu-west-2.pooler.supabase.com:5432/postgres
```

**Direct (for migrations / admin work):**
```
postgresql://postgres:[DB_PASSWORD]@db.mshugzgwjnyyjdkyjwgn.supabase.co:5432/postgres
```

## Local development

The Supabase project is managed from the `backend/` directory.

```bash
# Start local Supabase stack
cd backend
supabase start

# Link to the remote project (one-time)
supabase link --project-ref mshugzgwjnyyjdkyjwgn

# Push local migrations to remote
supabase db push

# Pull remote schema changes into a new migration file
supabase db pull <migration-name> --local --yes
```
