# Superhero API technical test

## Background

We are building a hypothetical service for customers, which streams their data to a superhero API. The API ingests information about superheroes, some of which is public, some must be kepy confidential. A schema for the expected structure of an API request's data [can be found in the resources directory](./resources/superHeroApiSpec.json).

Snowbridge is a tool which generically serves use cases to read data, optionally transform or filter it, and send it to a destination stream or http endpoint. [An overview of Snowbridge can be found in the documentation](https://docs.snowplow.io/docs/destinations/forwarding-events/snowbridge/).

The superhero service will use Snowbridge to transform input data, and send it to the superhero API. [Transformation is achieved via the javascript transformation script in the resources directory](./resources/superHeroAPI.js). Alongside [a sample Snowbridge configuration file](./resources/config.hcl). Snowbridge's configuration options [can be found in the documentation](https://docs.snowplow.io/docs/destinations/forwarding-events/snowbridge/configuration/), and additionally there is a short [guide to running the app locally](https://docs.snowplow.io/docs/destinations/forwarding-events/snowbridge/testing/).

There is a [sample of the input data for this service in the resources directory](./resources/input.txt).

## The Task

You are tasked with setting up testing to ensure that the service described above operates correctly, end-to-end. Assume that the features involved are already well covered by unit tests - the goal of this exercise is to provide coverage of the functionality of this specific service.

The main goal of this task is to establish a first iteration of testing, which can be easily expanded upon by other team members in the future. We expect to add to it regularly over time in small steps, and in the event that we run into an issue with the production service - when we need to replicate and respond quickly. 

We should assume that some of the team members adding to the tests, doing so may be their first interaction with this service, and some of them won't have much experience with Go.

Specific instructions:

- Use Go if possible (our internal QA setup is written using Go)
- Set up a test http server to receive the data in our tests
- Write a test/some tests to cover the core functionality of the service, as far as is practical
- Keep it simple
- Use code for as much of the process as is practical
- Write a short guide for how to run the the tests - a short readme or code comment would work, whatever is most fitting for your solution
- Comment your code as appropriate
- Add an example of a scenario which fails to conform to the API's expected structure

Push your solution to a repo and send us a link to it, rather than creating a PR in this repo.