//Tener en cuenta recibir 100 mediciones por segundo provenientes de sensores de diferentes tipos.
//El sistema no debe menguar el rendimiento, por ello y las consultas a realizarse no deben demorar más de 2 segundos en arrojar datos en escenarios límite.

import http from "k6/http";
import { check, sleep } from "k6";

// Options for the test
export const options = {
  vus: 100, // number of virtual users
  duration: "20s", // duration of the test
};

// Define the sensor configurations
const sensors = [
  {
    id: "1000",
    description: "Este es un sensor para la puerta",
    serialNumber: "The quick brown fox jumps over the lazy dog.2",
    brand: "Insano",
    address: "QCYO",
    reportStructure: {
      reports: [{ type: "Puerta", unit: "", value: "^(Abierta|Cerrada)$" }],
      sensorId: "1000",
    },
  },
  {
    id: "2000",
    description: "Este es un sensor para la ventana",
    serialNumber: "The quick brown fox jumps over the lazy dog.3",
    brand: "Insano",
    address: "WCYO",
    reportStructure: {
      reports: [
        {
          type: "Ventana",
          unit: "",
          value: "^(Abierta|Cerrada|Semi-abierta)$",
        },
      ],
      sensorId: "2000",
    },
  },
];

// Setup function to post sensor configurations
export function setup() {
  const url = "http://127.0.0.1:8090/sensor";
  for (const sensor of sensors) {
    const payload = JSON.stringify(sensor);
    const params = {
      headers: {
        "Content-Type": "application/json",
        "auth": "INSERT AUTH TOKEN HERE"
      },
    };

    const res = http.post(url, payload, params);

    // Log the response time
    console.log(`Sensor Post Response time: ${res.timings.duration} ms`);

    check(res, {
      "response time is less than 500ms": (r) => r.timings.duration < 500,
    });
  }
}

export default function () {
  // Define the report endpoint URL
  const url = "http://127.0.0.1:8090/reports/sensor";

  // Randomly select one of the sensors
  const sensor = sensors[Math.floor(Math.random() * sensors.length)];

  const payload = JSON.stringify({
    sensorID: sensor.id,
    date: "2020-10-01",
    reports: {
      [sensor.reportStructure.reports[0].type]: {
        value: Math.random() > 0.5 ? "Abierta" : "Cerrada",
        unit: "",
      },
    },
  });

  const params = {
    headers: {
      "Content-Type": "application/json",
    },
  };

  const res = http.post(url, payload, params);

  // Log the response time
  console.log(`Report Post Response time: ${res.timings.duration} ms`);

  check(res, {
    "status is 200": (r) => r.status === 200,
    "response time is less than 1000ms": (r) => r.timings.duration < 1000,
    "response body is not empty": (r) => r.body.length > 0,
  });

  sleep(0.01); // wait for 10 milliseconds between requests
}
