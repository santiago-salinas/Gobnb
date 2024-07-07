import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  vus: 100, // The number of virtual users
  duration: "1m", // Test duration
  thresholds: {
    http_req_duration: ["p(95)<2000"], // 95% of requests should be below 2 seconds
  },
};

const BASE_URL = "http://127.0.0.1:8090";
const TOKEN = "test_token_one";
const propertyId = "qy40nbutxtxlpcx";

const headers = {
  "Content-Type": "application/json",
  auth: TOKEN,
};

export function setup() {
  // Pay for the property if it ain't paid
  const paymentPayload = JSON.stringify({
    propertyId,
    cardInfo: {
      cardNumber: "1234567812345678",
      name: "Ruperto Rocanrol",
      cvv: "123",
      expDate: "2025-06",
    },
  });

  http.post(`${BASE_URL}/property/pay`, paymentPayload, { headers });

  return { propertyId };
}

export default function () {
  const EMAILTOKEN = generateRandomEmail();
  const randomDay = generateRandomDate();
  const reservationPayload = JSON.stringify({
    document: "TrustMe",
    name: "Santiago",
    last_name: "Salinas",
    email: EMAILTOKEN,
    phone: "+598 1234567",
    address: "Mongo 123",
    nationality: "Uy",
    country: "UY",
    adults: 1,
    minors: 0,
    property: propertyId,
    reserved_from: randomDay,
    reserved_until: randomDay,
  });

  const reservationRes = http.post(
    `${BASE_URL}/reservations`,
    reservationPayload,
    { headers }
  );
  check(reservationRes, {
    "reservation status is 201": (r) => r.status === 201,
    "response time is less than 500ms": (r) => r.timings.duration < 500,
  });

  const viewReservationRes = http.get(
    `${BASE_URL}/reservations?email=${EMAILTOKEN}`,
    { headers }
  );
  const reservations = viewReservationRes.json();
  let reservationId = null;

  for (let i = 0; i < reservations.length; i++) {
    if (
      reservations[i].email === EMAILTOKEN &&
      reservations[i].property === propertyId
    ) {
      reservationId = reservations[i].id;
      break;
    }
  }

  if (reservationId) {
    console.log(`Reservation ID: ${reservationId}`);
  } else {
    console.log("No reservation found with the given email and property ID");
  }

  // Approve the reservation
  const approvalRes = http.post(
    `${BASE_URL}/reservations/${reservationId}/approve`,
    {},
    { headers }
  );
  check(approvalRes, {
    "reservation approval status is 200": (r) => r.status === 200,
    "response time is less than 500ms": (r) => r.timings.duration < 500,
  });

  // Pay for the reservation
  const reservationPaymentPayload = JSON.stringify({
    reservationId: reservationId,
    cardInfo: {
      cardNumber: "1234567812345678",
      name: "Ruperto Rocanrol",
      cvv: "123",
      expDate: "2025-06",
    },
  });

  const reservationPaymentRes = http.post(
    `${BASE_URL}/reservations/pay`,
    reservationPaymentPayload,
    { headers }
  );
  check(reservationPaymentRes, {
    "reservation payment status is 201": (r) => r.status === 201,
    "response time is less than 500ms": (r) => r.timings.duration < 500,
  });

  const removeRes = http.post(
    `${BASE_URL}/reservations/${reservationId}/remove`,
    {},
    { headers }
  );
  check(removeRes, {
    "reservation remove status is 200": (r) => r.status === 200,
    "response time is less than 500ms": (r) => r.timings.duration < 500,
  });

  // Short sleep to ensure we reach 1000 requests in the given duration
  sleep(1);
}

function generateRandomEmail() {
  const domains = ["gmail.com", "yahoo.com", "outlook.com", "example.com"];
  const chars = "abcdefghijklmnopqrstuvwxyz1234567890";

  let emailPrefix = "";
  for (let i = 0; i < 10; i++) {
    emailPrefix += chars[Math.floor(Math.random() * chars.length)];
  }

  const domain = domains[Math.floor(Math.random() * domains.length)];

  return `${emailPrefix}@${domain}`;
}

function generateRandomDate() {
  const startYear = 2025;
  const endYear = 9999;

  // Generate a random year between startYear and endYear
  const year =
    Math.floor(Math.random() * (endYear - startYear + 1)) + startYear;

  // Generate a random month between 1 and 12
  const month = Math.floor(Math.random() * 12) + 1;

  // Generate a random day based on the month
  let day;
  if (month === 2) {
    // Check for leap year
    const isLeapYear = (year % 4 === 0 && year % 100 !== 0) || year % 400 === 0;
    day = Math.floor(Math.random() * (isLeapYear ? 29 : 28)) + 1;
  } else if ([4, 6, 9, 11].includes(month)) {
    day = Math.floor(Math.random() * 30) + 1;
  } else {
    day = Math.floor(Math.random() * 31) + 1;
  }

  // Format the month and day with leading zeros if necessary
  const formattedMonth = month.toString().padStart(2, "0");
  const formattedDay = day.toString().padStart(2, "0");

  return `${year}-${formattedMonth}-${formattedDay}`;
}
