# skAIrocKet - Backend
## Project Introduction
+ The main objective of this project is to develop a tool capable of predicting stocks with potential for rapid growth, and to provide a web UI for user interaction. Users can input the desired stock number into the web UI, and the website will display the predicted next-day stock price along with the confidence score for the prediction.   
+ If users have registered accounts and logged in, they can choose to add their favorite stocks for convenience in daily checking. User information is managed using MySQL and MongoDB databases.
    
+ For user authentication, this project utilizes the Refresh Token framework to provide security assurance.
    
+ Regarding the model dataset, I used the constituents of the Taiwan Stock ETF 00733 from four quarters as the dataset. This ETF selects high momentum small and medium-sized stocks calculated by a formula. I believe that these selected stocks share similarities in data distribution, hence I used them as the dataset.
    
+ I trained two models for prediction. The first is an LSTM regression prediction model, which uses the stock information of the previous ten days, including opening price, closing price, trading volume, etc., to predict the closing price of the next day. The other is an LSTM auto-encoder model, which also uses ten days of data as input. It observes the similarity between the output and input, and calculates the prediction confidence using the reconstruction error of the model output. A lower reconstruction error indicates that the input data has similar data distribution to the training data, thus giving a higher prediction confidence score.
    
+ The frontend of the website is developed using the Vue framework along with the Bootstrap-Vue package to create a pleasing user interface. Axios is used for communication with the backend using REST API.
    
+ Finally, the entire project is packaged into Docker images using the two-stage build method to produce a smaller image. Docker Compose files is used to simplify the deployment process. With Docker configured on any machine, the entire website deployment can be completed with simple commands.
 

## Tech Stack
### Backend
+ Golang 1.19
+ Refresh Token Authentication
### Deep Learning Models
+ LSTM Regression Model
+ LSTM Auto-Encoder Model
### Frontend
+ Vue.js 2.7.16
+ Bootstrap-vue
### Deployment
+ Microservice Architecture using Docker

### Ongoing Sections
+ HTTP Response exception handling
+ Standardize variable names