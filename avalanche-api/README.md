# Avalanche API

## Context

Snowplow pipeline is a high-throughput system processing billions of events. The API you are working on is a critical part of the wider ecosystem. Therefore any additonal latency, instability affects it directly. The service is called within an enrichment process that attaches additional data to each event. This means thousands of events within each request flow through the system almost every second.

The system provides additional information based upon historically aggregated attributes and profiles.

Additional context on attribute processing:
- Each attribute is assigned to at least one profile.
- If any attribute within an event is already linked to an exising profile, all the attributes in that event need to be linked to the profile.
- If an attribute is assigned to many profiles, the oldest one is the one to use as a response.
- We explicitly don't want to follow many hops between attributes and profiles. We only care about direct attribute to profile connections.

## Task

> [!IMPORTANT]
> Push your solution to a **private GitHub repo** and invite us to it.  
> Add the following GitHub users: **peel, colmsnowplow, istreeter, jbeemster, AlexBenny**.  

> [!WARNING]
> Please, **do not create a PR against this repo**.  

### Non-deterministic id resolution (algorithmic)
Currently, when a large batch of events is received even if some of the events would normally be attributed to the same profile, they get different ids. We need to make sure all related events on a batch get a determinsitic id.

Given, following POST body, because the events share no attributes they will get different ids.
```
POST :endpoint/events
:headers

[
  { "eventId": "10",  "userId": "10", "email": "user10@snowplow.io" },
  { "eventId": "11",  "userId": "11", "email": "user11@snowplow.io" },
  { "eventId": "12",  "userId": "12", "email": "user12@snowplow.io" },
  { "eventId": "13",  "userId": "13", "email": "user13@snowplow.io" },
  { "eventId": "14",  "userId": "14", "email": "user14@snowplow.io" },
  { "eventId": "15",  "userId": "15", "email": "user15@snowplow.io" }
]
```
<details>
<summary>curl command</summary>

```shell
curl -i -H User-Agent\:\ SomeApp/1.0 -H Content-Type\:\ application/json -XPOST http\://localhost\:9191/events -d \['
'\ \ \{\ \"eventId\"\:\ \"10\"\,\ \ \"userId\"\:\ \"10\"\,\ \"email\"\:\ \"user10\@snowplow.io\"\ \}\,'
'\ \ \{\ \"eventId\"\:\ \"11\"\,\ \ \"userId\"\:\ \"11\"\,\ \"email\"\:\ \"user11\@snowplow.io\"\ \}\,'
'\ \ \{\ \"eventId\"\:\ \"12\"\,\ \ \"userId\"\:\ \"12\"\,\ \"email\"\:\ \"user12\@snowplow.io\"\ \}\,'
'\ \ \{\ \"eventId\"\:\ \"13\"\,\ \ \"userId\"\:\ \"13\"\,\ \"email\"\:\ \"user13\@snowplow.io\"\ \}\,'
'\ \ \{\ \"eventId\"\:\ \"15\"\,\ \ \"userId\"\:\ \"15\"\,\ \"email\"\:\ \"user15\@snowplow.io\"\ \}'
'\]'
''
'
```

</details>

On the other hand, the following request is sent, all of the events should get the same id, because they share at least one attribute:

```
POST :endpoint/events
:headers

[
  { "eventId": "10",  "userId": "10", "email": "user10@snowplow.io" },
  { "eventId": "11",  "userId": "10", "email": "user11@snowplow.io" },
  { "eventId": "12",  "userId": "12", "email": "user10@snowplow.io" },
  { "eventId": "13",  "userId": "10", "email": "user13@snowplow.io" },
  { "eventId": "14",  "userId": "14", "email": "user10@snowplow.io" },
  { "eventId": "15",  "userId": "10", "email": "user15@snowplow.io" }
]
```

<details>
<summary>curl command</summary>

```shell
curl -i -H User-Agent\:\ SomeApp/1.0 -H Content-Type\:\ application/json -XPOST http\://localhost\:9191/events -d \['
'\ \ \{\ \"eventId\"\:\ \"100\"\,\ \ \"userId\"\:\ \"100\"\,\ \"email\"\:\ \"user100\@snowplow.io\"\ \}\,'
'\ \ \{\ \"eventId\"\:\ \"110\"\,\ \ \"userId\"\:\ \"100\"\,\ \"email\"\:\ \"user110\@snowplow.io\"\ \}\,'
'\ \ \{\ \"eventId\"\:\ \"120\"\,\ \ \"userId\"\:\ \"120\"\,\ \"email\"\:\ \"user100\@snowplow.io\"\ \}\,'
'\ \ \{\ \"eventId\"\:\ \"130\"\,\ \ \"userId\"\:\ \"100\"\,\ \"email\"\:\ \"user130\@snowplow.io\"\ \}\,'
'\ \ \{\ \"eventId\"\:\ \"140\"\,\ \ \"userId\"\:\ \"140\"\,\ \"email\"\:\ \"user130\@snowplow.io\"\ \}\,'
'\ \ \{\ \"eventId\"\:\ \"150\"\,\ \ \"userId\"\:\ \"100\"\,\ \"email\"\:\ \"user150\@snowplow.io\"\ \}\,'
'\ \ \{\ \"eventId\"\:\ \"160\"\,\ \ \"userId\"\:\ \"160\"\,\ \"email\"\:\ \"user150\@snowplow.io\"\ \}'
'\]'
'
```

</details>

It takes several times with the same request for all of them to return the same id.


### Specific instructions

- Implement a solution to the described problem
- If necessary write a summary of the approach and how to verify it
- Write tests to cover the functionality and prove that it works
- Think about metrics that could be used to monitor the feature - can we detect when it's misbehaving or the input data makes it problematic?
- Try to understand the performance semantics of the app - can you spot anything odd?
- What if we had to change the API - how can we prevent breaking the contract?

We will discuss the implementation and the approach taken.
