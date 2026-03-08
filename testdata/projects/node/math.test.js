import { expect, test } from "vitest";
import { add, isPositive } from "./math.js";

test("add", () => {
  expect(add(2, 3)).toBe(5);
});

test("isPositive", () => {
  expect(isPositive(1)).toBe(true);
  expect(isPositive(-1)).toBe(false);
  expect(isPositive(0)).toBe(false);
});
