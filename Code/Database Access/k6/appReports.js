import http from "k6/http";
import { check, sleep } from "k6";

// Options for the test
export const options = {
  vus: 100, // number of virtual users
  duration: "1s", // duration of the test
};

// Define the sensor configurations
const properties = ["676", "xve9391o2j1ls5p", "eh8cnak5rgha7dp"];

// Define possible report types and values
const reportTypes = [
  { type: "Llaves", values: ["No estan", "Perdidas", "En la oficina"] },
  { type: "Puerta", values: ["Abierta", "Cerrada", "Atascada"] },
  { type: "Ventana", values: ["Abierta", "Cerrada", "Rota"] },
  { type: "Luz", values: ["Encendida", "Apagada", "Intermitente"] },
  { type: "Temperatura", values: ["Alta", "Normal", "Baja"] }
];

// Utility function to get a random date in 2024
function getRandomDate2024() {
  const start = new Date("2024-01-01");
  const end = new Date("2024-12-31");
  return new Date(start.getTime() + Math.random() * (end.getTime() - start.getTime())).toISOString().split('T')[0];
}

export default function () {
  // Define the report endpoint URL
  const url = "http://127.0.0.1:8090/reports/app";

  // Randomly select one of the properties
  const propertyId = properties[Math.floor(Math.random() * properties.length)];

  // Randomly select a report type and value
  const reportType = reportTypes[Math.floor(Math.random() * reportTypes.length)];
  const reportValue = reportType.values[Math.floor(Math.random() * reportType.values.length)];

  // Generate a random sensor ID starting with "APP"
  const sensorID = `APP${Math.floor(Math.random() * 10000).toString().padStart(4, '0')}`;

  // Generate a random date in 2024
  const date = getRandomDate2024();

  const payload = JSON.stringify({
    sensorID: sensorID,
    date: date,
    type: reportType.type,
    value: reportValue,
    propertyId: propertyId,
  });

  const params = {
    headers: {
      "Content-Type": "application/json",
    },
  };

  const res = http.post(url, payload, params);

  // Log the response time
  console.log(`App Report Post Response time: ${res.timings.duration} ms`);

  check(res, {
    "status is 200": (r) => r.status === 200,
    "response time is less than 1000ms": (r) => r.timings.duration < 1000,
    "response body is not empty": (r) => r.body.length > 0,
  });

  if (res.status !== 200) {
    console.error(`Failed to post app report for propertyId: ${propertyId}`);
  }

  sleep(0.01); // wait for 10 milliseconds between requests
}
