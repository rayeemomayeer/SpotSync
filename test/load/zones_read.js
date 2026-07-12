# SpotSync load — zone list read mix

import http from "k6/http";
import { check, sleep } from "k6";

const BASE = __ENV.API_BASE || "http://localhost:8080/api/v1";

export const options = {
  vus: 20,
  duration: "30s",
  thresholds: {
    http_req_failed: ["rate<0.05"],
    http_req_duration: ["p(95)<2000"],
  },
};

export default function () {
  const res = http.get(`${BASE}/zones`);
  check(res, {
    "zones 200": (r) => r.status === 200,
    "envelope ok": (r) => {
      try {
        return JSON.parse(r.body).success === true;
      } catch {
        return false;
      }
    },
  });
  sleep(0.5);
}
