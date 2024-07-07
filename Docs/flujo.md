Guia de como Crear un inmueble, subir las fotos, pagarlo, reservarlo, aprobar la reserva, pagar la reserva

```
POST http://127.0.0.1:8090/property
```
```
HEADERS auth {{token}}
```
```
{
  "name": "La casa de mickey mouse",
  "adultQuantity": 5,
  "kidQuantity": 2,
  "kingSizedBeds": 3,
  "singleBeds": 2,
  "hasAC": "false",
  "hasWIFI": "true",
  "hasGarage": "true",
  "type": 2,
  "beachDistance": 5000,
  "state": "Florida",
  "resort": "Miami",
  "neighborhood": "Disney",
  "bookingPrice": 10,
  "unavailableDates": [
    {
      "start": "2024-12-20",
      "end": "2025-01-05"
    }
  ]
}
```
![image](https://github.com/IngSoft-AR-2023-2/266628_271568_255981/assets/48341470/0448edfd-7021-4cd6-858e-32872580aeac)

En este estado, estará sin pagar, y no estará pendiente de pago. Se deben subir las 4 fotos para que esté pendiente de pago.

![image](https://github.com/IngSoft-AR-2023-2/266628_271568_255981/assets/48341470/a73807c4-914c-494f-8dae-c45742ff35fa)


Luego se postea de a una, las fotos. 4 veces
```
POST http://127.0.0.1:8090/property/img/vsy9nnhyurmg9vu
```
```
HEADERS auth {{token}}
```
![image](https://github.com/IngSoft-AR-2023-2/266628_271568_255981/assets/48341470/789b2eaf-ffab-4358-9478-eef11891184c)

Ahora la propiedad estará pendiente de pago

![image](https://github.com/IngSoft-AR-2023-2/266628_271568_255981/assets/48341470/a77a26df-5d4a-454a-bcc1-8a7fb4954f1c)

```
POST http://127.0.0.1:8090/property/pay
```
```
{
    "propertyId": "vsy9nnhyurmg9vu",
    "cardInfo": {
        "cardNumber": "1234567812345678",
        "name": "Ruperto Rocanrol",
        "cvv": "123",
        "expDate": "2025-06"
    }
}
```
![image](https://github.com/IngSoft-AR-2023-2/266628_271568_255981/assets/48341470/6f9c9f1f-dc07-46e9-a8b6-2ceca0d551c2)

Ahora el property se encuentra en el siguiente estado

![image](https://github.com/IngSoft-AR-2023-2/266628_271568_255981/assets/48341470/b68804f0-7f71-44f7-bcc5-a2fce52d5409)

Y ahora es capaz de aparecer en las búsquedas

![image](https://github.com/IngSoft-AR-2023-2/266628_271568_255981/assets/48341470/99830005-6471-4b14-aeaa-adf0a3788467)

Incluso podemos acceder a las fotos

![image](https://github.com/IngSoft-AR-2023-2/266628_271568_255981/assets/48341470/d7fa59a9-6c49-4a5d-836d-82eea3a35f5d)


Vamos a hacer una reserva
```
POST http://127.0.0.1:8090/reservations
```
```
  {
    "document": "TrustMe",
    "name": "Santiago",
    "last_name": "Salinas",
    "email": "santiagosalinasoto@gmail.com",
    "phone": "+598 1234567",
    "address": "Mongo 123",
    "nationality": "Uy",
    "country": "UY",
    "adults": 1,
    "minors":0,
    "property": "vsy9nnhyurmg9vu",
    "reserved_from": "2024-12-01",
    "reserved_until": "2024-12-30"
  }
```

Se debe aprobar la reserva, pero debemos saber el id de reserva usando
```
GET http://127.0.0.1:8090/reservations
```

Mientras tanto estará pending

![image](https://github.com/IngSoft-AR-2023-2/266628_271568_255981/assets/48341470/2f7bca89-aeff-426a-924c-d09422925333)

```
POST http://127.0.0.1:8090/reservations/60use7iqdk0ijt9/approve
```
![image](https://github.com/IngSoft-AR-2023-2/266628_271568_255981/assets/48341470/a048ba34-dd7a-4c58-b715-ec5b356bd598)

```
POST http://127.0.0.1:8090/reservations/pay
```
```
{
    "reservationId": "60use7iqdk0ijt9",
    "cardInfo": {
        "cardNumber": "1234567812345678",
        "name": "Ruperto Rocanrol",
        "cvv": "123",
        "expDate": "2025-06"
    }
}
```

Cancelar reserva
```
POST http://127.0.0.1:8090/reservations/60use7iqdk0ijt9/cancel
```
![image](https://github.com/IngSoft-AR-2023-2/266628_271568_255981/assets/48341470/bfcf0a93-2314-4287-a093-363bcdf44217)

Nota: No se puede volver a reservar un lugar, con el mismo email



En K6 Hicimos un test donde por un minuto haciamos este proceso 2000 veces. En average el request demora 410ms
![image](https://github.com/IngSoft-AR-2023-2/266628_271568_255981/assets/48341470/c421a3e7-b00b-43b7-9872-11b6c5760d79)

