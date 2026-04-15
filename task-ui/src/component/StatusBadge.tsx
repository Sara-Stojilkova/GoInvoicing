export type Status = "todo" | "in_progress" | "done";

const statusBadge: Record<Status, string> = {
  todo: "badge badge-gray",
  in_progress: "badge badge-yellow",
  done: "badge badge-green",
};

const statusLabel: Record<Status, string> = {
  todo: "Todo",
  in_progress: "In Progress",
  done: "Done",
};

export function StatusBadge({ status }: { status: Status }) {
  return (
    <span className={statusBadge[status] ?? "badge badge-gray"}>
      {statusLabel[status] ?? status}
    </span>
  );
}