# Portfolio-Tools
[![GoDoc](https://godoc.org/github.com/jakeschurch/portfolio-tools?status.svg)](https://godoc.org/github.com/jakeschurch/portfolio-tools)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://opensource.org/licenses/MIT)

## Basic Overview
Portfolio-Tools is a package that allows for the storage of information of a particular security.

### Attributes to Implement
0. Ticker string
1. Historical Price Data **TBD**
2. Date Bought datetime
3. Date Sold datetime
4. Quantity *(int or float64?)*
5. Order Type [Buy, Sell, Limit]

## Historical Data Slices

**Many ways that we can store data. Just need to figure out best way...**
QUESTION: How detailed do we want to store data? Are we storing data at tick-level or the day-level?

### 0. Object-Oriented Route
- Attributes:
    - Date datetime
    - Price float64
    - Volume float64
- Implementation:
    - Create class instance
    - Store all instances in slice index
- Easy to access slice element, reference by address


### 1.  Jagged Slice Route
- Access elements by slice index
- Most likely faster than implementation #0
- reference elements by value
#### Example Implementation:
![](https://latex.codecogs.com/png.latex?%5Cinline%20%5Cbegin%7Bbmatrix%7D%20%5Cbegin%7Bbmatrix%7D%20Datetime_0%2C%20Price_0%2C%20Volume_0%2C%20%5Cend%7Bbmatrix%7D%2C%20%5C%5C%20%5Cbegin%7Bbmatrix%7D%20Datetime_1%2C%20Price_1%2C%20Volume_1%2C%20%5Cend%7Bbmatrix%7D%2C%20%5C%5C%20%5Cvdots%20%5C%5C%20%5Cbegin%7Bbmatrix%7D%20Datetime_n%2C%20Price_n%2C%20Volume_n%2C%20%5Cend%7Bbmatrix%7D%2C%20%5Cend%7Bbmatrix%7D)
#### 1a. Further Considerations:
 - Potentially utilize hashmap that records all potential securities
     - From there, record transactions in a particular class instance
