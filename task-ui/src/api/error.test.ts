import { describe, it, expect } from "vitest";
import { ApiError } from "./error";

describe("ApiError", () => {
  it("stores the status code", () => {
    const err = new ApiError(404, "not found");
    expect(err.status).toBe(404);
  });

  it("stores the message", () => {
    const err = new ApiError(500, "internal server error");
    expect(err.message).toBe("internal server error");
  });

  it("is an instance of Error", () => {
    const err = new ApiError(400, "bad request");
    expect(err).toBeInstanceOf(Error);
  });

  it("is an instance of ApiError", () => {
    const err = new ApiError(401, "unauthorized");
    expect(err).toBeInstanceOf(ApiError);
  });

  it("can be caught as a plain Error", () => {
    const err = new ApiError(403, "forbidden");
    expect(() => { throw err; }).toThrow(Error);
  });
});
