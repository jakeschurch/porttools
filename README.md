# Go-Security Pkg Documentation
##### Version: 0.0.1
---
Go-Security is a package that allows for the storage of information of a particular security.
**TBD: Go-Transaction(?)
TODO: Go-Portfolio**

### Attributes to Implement
---
0. Ticker string
1. Historical Price Data **TBD** 
2. Date Bought datetime
3. Date Sold datetime
4. Quantity *(int or float64?)*
5. Order Type [Buy, Sell, Limit] 

## Historical Data Slices
---
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
$$\begin{bmatrix} 
\begin{bmatrix}
Datetime_0, Price_0, Volume_0,
\end{bmatrix}, \\\\
\begin{bmatrix}
Datetime_1, Price_1, Volume_1,
\end{bmatrix}, \\\\
 \vdots \\\\
\begin{bmatrix}
Datetime_n, Price_n, Volume_n,
\end{bmatrix},
\end{bmatrix}$$

#### 1a. Further Considerations:
 - Potentially utilize hashmap that records all potential securities
     - From there, record transactions in a particular class instance

## TODO: Methods
---
0. Get Last (latest) price
1. String Representation (for logging)




