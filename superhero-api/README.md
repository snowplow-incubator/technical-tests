# Superhero API technical test

## Background

We are building a hypothetical service for customers, which streams their data to a superhero API. The API ingests information about superheroes, keeping some information confidential, and some of it public. A schema for the expected structure of an API request's data [can be found in the resources directory](./resources/superHeroApiSpec.json).

Snowbridge is a tool which generically serves use cases to read data, optionally transform or filter it, and send it to a destination stream or http endpoint. [An overview of Snowbridge can be found in the documentation](https://docs.snowplow.io/docs/destinations/forwarding-events/snowbridge/).

The superhero service will use Snowbridge to transform input data, and send it to the superhero API. Transformation is achieved via the [javascript transformation script in the resources directory](./resources/superHeroAPI.js). Alongside [a sample Snowbridge configuration file](./resources/config.hcl). Snowbridge's configuration options [can be found in the documentation](https://docs.snowplow.io/docs/destinations/forwarding-events/snowbridge/configuration/), additionally there is a short [guide to running the app locally](https://docs.snowplow.io/docs/destinations/forwarding-events/snowbridge/testing/).

There is a [sample of input data for this service in the resources directory](./resources/input.txt).

## The Task

You are tasked with setting up testing to ensure that the service described above operates correctly, from entry to Snowbridge, through to reaching the api. We do not have access to a test API, but we must aim to confirm that we will correctly make requests to the API, insofar as possible. Assume that the features of Snowbridge are already well covered by unit tests - the goal of this exercise is to provide coverage of the functionality of this specific service.

The main aim is to establish a first iteration of testing - which covers the most important core funcionality, and can be easily expanded upon by other team members in the future. We expect to add to it regularly over time in small steps. We also wish to use and add to this setup in the event that we run into an issue with the production service - when we need to replicate and respond quickly.

We should assume that some of the team members adding to the tests, doing so may be their first interaction with this service, and some of them may not have much experience with Go.

Specific instructions:

- Use Go if possible (our internal QA setup is written using Go)
- Set up a test http server to receive the data in our tests
- Write a test/some tests to cover the core functionality of the service, as far as is practical
- Keep it simple
- Use code for as much of the process as is practical
- Write a short guide for how to run the the tests - a short readme or code comment would work. Choose whatever is most fitting for your solution
- Comment your code as appropriate
- Add an example of a scenario which fails to conform to the API's expected data structure

Push your solution to a repo and send us a link to it, rather than creating a PR in this repo.