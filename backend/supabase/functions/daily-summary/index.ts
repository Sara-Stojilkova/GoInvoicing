// @ts-nocheck — Deno runtime; use the Deno VS Code extension for full type checking
import { createClient } from "@supabase/supabase-js";

Deno.serve(async (req: Request) => {
  const supabase = createClient(
    Deno.env.get("SUPABASE_URL")!,
    Deno.env.get("SUPABASE_SERVICE_ROLE_KEY")!
  );

  // Calculate yesterday range (UTC)
  const today = new Date();
  const start = new Date(Date.UTC(
    today.getUTCFullYear(),
    today.getUTCMonth(),
    today.getUTCDate() - 1
  ));
  const end = new Date(Date.UTC(
    today.getUTCFullYear(),
    today.getUTCMonth(),
    today.getUTCDate()
  ));

  // Query tasks completed yesterday
  const { data, error } = await supabase
    .from("tasks")
    .select("agency_id")
    .eq("status", "done")
    .gte("completed_at", start.toISOString())
    .lt("completed_at", end.toISOString());

  if (error) {
    return new Response(JSON.stringify({ error: error.message }), {
      status: 500,
      headers: { "Content-Type": "application/json" },
    });
  }

  // Count per agency
  const counts: Record<string, number> = {};

  for (const row of data) {
    counts[row.agency_id] = (counts[row.agency_id] || 0) + 1;
  }

  return new Response(JSON.stringify(counts), {
    headers: { "Content-Type": "application/json" },
  });
});
