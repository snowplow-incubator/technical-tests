# Superhero API technical test

## Background

We are building a new service for our customers: It is called **The Superhero Service**.

Our customers like to collect data about superheroes.  Our customers provide to us a source of Superhero data which they have collected.  Our new Superhero Service must stream this data from the customer into a Superhero HTTP API. The HTTP API ingests information about superheroes, keeping some information confidential, and some of it public. A schema for the expected structure of an API request's data [can be found in the resources directory](./resources/superHeroApiSpec.json).

<< Insert picture here showing source of data --> superhero service --> superhero api >>

We are implementing The Superhero Service using [Snowbridge](https://docs.snowplow.io/docs/destinations/forwarding-events/snowbridge/). Snowbridge is a general-purpose streaming tool which can read data, transform or filter the data, and send the data to a destination http endpoint. [An overview of Snowbridge can be found in the documentation](https://docs.snowplow.io/docs/destinations/forwarding-events/snowbridge/).

For the implementation of The Superhero Service, we have created some files which you can find in the `resources` directory:

- A [sample of a customer's input data](./resources/input.txt).
- A [javascript transformation script](./resources/superHeroAPI.js) which will transform the customer's input data into a format that can be ingested by the superhero API.
- A [Snowbridge configuration file](./resources/config.hcl) which tells Snowbridge how to read, transform, and write the source of data.

And that's all!  Those files are all that is needed to run The Superhero Service via Snowbridge.

Snowbridge's configuration options [can be found in the documentation](https://docs.snowplow.io/docs/destinations/forwarding-events/snowbridge/configuration/), additionally there is a short [guide to running the app locally](https://docs.snowplow.io/docs/destinations/forwarding-events/snowbridge/testing/).

## Your Task

**Your task is to set up testing to ensure that The Superhero Service operates correctly.**

Your test must demonstrate that The Superhero Service correctly reads the source of data, transforms the data, and forwards the data to the Superhero HTTP API.

Unfortunately, for our tests we are not allowed to use the _real_ Superhero HTTP API servers.  But you must still aim to validate how the Superhero Service sends data to a HTTP endpoint, insofar as possible.

Assume that the features of Snowbridge are already well covered by unit tests; you are not required to demonstrate the general features of Snowbridge. The goal of this exercise is to test the specific functionality of The Superhero Service.

Specific instructions:

- Use Go if possible (our internal QA setup is written using Go)
- Set up a test HTTP server to receive the data in your tests
- Write tests to cover the core functionality of The Superhero Service, as far as is practical
- Keep it simple
- Use code for as much of the process as is practical
- Write a short guide for how to run the the tests - a short readme or code comment would work. Choose whatever is most fitting for your solution
- ~Comment your code as appropriate~
- ~Add an example of a scenario which fails to conform to the API's expected data structure~

Push your solution to a GitHub repo and send us a link to it. Please do not create a PR against this repo.



#### I would cut these lines:

The main aim is to establish a first iteration of testing - which covers the most important core funcionality, and can be easily expanded upon by other team members in the future. We expect to add to it regularly over time in small steps. We also wish to use and add to this setup in the event that we run into an issue with the production service - when we need to replicate and respond quickly

We should assume that some of the team members adding to the tests, doing so may be their first interaction with this service, and some of them may not have much experience with Go.

