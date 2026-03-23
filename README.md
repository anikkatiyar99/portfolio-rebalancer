# Portfolio Rebalancer

This is a take-home assignment to build backend APIs for managing and rebalancing user portfolios.


## Tech Stack

- Go
- Elasticsearch (Feel free to use any other database or an in-memory alternative)
- Kafka (Feel free to use any other messaging system if needed)
- Docker


## Running the Project

```
docker compose build
docker compose up
```

API base URL: `http://localhost:8080`

Swagger UI: `http://localhost:8080/docs/index.html`

OpenAPI JSON: `http://localhost:8080/docs/doc.json`


## Models

- Portfolio 
        - `UserID` field is a unique user identifier in our system
        - `Allocation` field represents the percentage of the user's total portfolio or cash distribution across different asset classes. 
            Eg: {"stocks": 60, "bonds": 30, "gold": 10}.
            Note: This means 60% of the user's portfolio is allocated to stocks, 30% to bonds, and 10% to gold
            
- UpdatedPortfolio 
        - `UserID` is the user's unique ID
        - `NewAllocation` is the new allocation of user portfolio in %.

- RebalanceTransaction
        - `userID` is the user's unique ID
        - `Action` is the type of transaction (BUY/SELL)
        - `Asset` is the type of user asset to be transferred (eg: stocks, bonds, gold etc.)
        - `RebalancePercent` is the percentage of the asset transferred

- Feel free to edit/add models


## APIs
- /portfolio : This takes in userId and current user allocation. This will api will be used to create users in our system along with their portfolio allocation.

- /rebalance : This is the API that simulates a third-party provider, which calculates a user's portfolio allocation based on market changes and returns an updated allocation. For the current task, we will manually call this API to mock the third-party interaction.


- Feel free to edit/add APIs


## TODO

- Complete the `/portfolio` API
    - Accept a new user's portfolio details via a POST request.
    - Persist the portfolio in Elasticsearch.

- Complete the `/rebalance` API
    - Accept a user's updated portfolio based on market conditions via a POST request.
    - Maintain the user's original allocation percentages for reference.
    - Calculate the transactions needed to rebalance the user's current portfolio allocation percentage back to their original allocation percentage.
    - Save the RebalanceTransaction in Elasticsearch.

- Assuming we could get multiple rebalance api calls from the provider, we need to ensure our system can handle load and is fault tolerant(could be supported by adding queue and retries).

- Write a README

- Feel free to add further capabilities.


## Example

- `/portfolio` API creates a user with ID = 1 and Allocation = {"stocks": 60, "bonds": 30, "gold": 10}
    Note: here the allocation is 60% stocks, 30% bonds and 10% gold

- `/rebalance` API is called with inputs
    ID = 1
    NewAllocation = {"stocks": 70, "bonds": 20, "gold": 10}
    [This is how much the user's portfolio has moved due to market conditions]

- We need to calculate and save the RebalanceTransaction to maintain 60% stocks, 30% bonds and 10% gold.

    Transaction 1: 
            UserID = "1"
	        Sell 10% of stocks

    Transaction 2: 
            UserID = "1"
	        Buy 10% of bonds


## Evaluation Criteria

- Code quality and structure
- Logical correctness
- Fault tolerance
- Extensibility and Scalablility
- Test coverage
- Optional: Error handling and edge cases.
