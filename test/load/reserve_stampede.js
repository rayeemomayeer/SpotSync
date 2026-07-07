import http from "k6/http";
import { check, sleep } from "k6";

const BASE = __ENV.API_BASE || "http://localhost:8080/api/v1";
const ZONE_ID = __ENV.ZONE_ID || "1";

export const options = {
  vus: 50,
  duration: "10s",
  thresholds: {
    http_req_failed: ["rate<0.5"],
    checks: ["rate>0.5"],
  },
};

export default function () {
  const email = `load-${__VU}-${__ITER}@test.local`;
  const reg = http.post(
    `${BASE}/auth/register`,
    JSON.stringify({
      name: "Load User",
      email,
      password: "password123",
      role: "driver",
    }),
    { headers: { "Content-Type": "application/json" } },
  );
  check(reg, { "register ok": (r) => r.status === 201 || r.status === 409 });

  const login = http.post(
    `${BASE}/auth/login`,
    JSON.stringify({ email, password: "password123" }),
    { headers: { "Content-Type": "application/json" } },
  );
  const token = login.json("data.token");
  if (!token) {
    sleep(0.1);
    return;
  }

  const res = http.post(
    `${BASE}/reservations`,
    JSON.stringify({ zone_id: Number(ZONE_ID), license_plate: `L${__VU}${__ITER}`.slice(0, 15) }),
    {
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
    },
  );
  check(res, {
    "reserve not 500": (r) => r.status !== 500,
    "reserve ok or full": (r) => r.status === 201 || r.status === 409,
  });
  sleep(0.05);
}
