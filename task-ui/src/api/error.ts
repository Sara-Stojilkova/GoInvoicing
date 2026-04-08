export class ApiError extends Error {
  status: number;

  constructor(_status: number, _message: string) {
    throw new Error("not implemented");
  }
}
