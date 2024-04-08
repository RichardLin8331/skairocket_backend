from FinMind.data import DataLoader
import yaml
from flask import Flask, render_template, request, Response
import pandas as pd
import json
import requests
import numpy as np
from numpyencoder import NumpyEncoder
import joblib

# Load config
config_path = './configs/stock_info.yaml'
global data_api
global stockinfo_config

month_start_end = [ ['12-01', '12-31'], ['01-01', '01-31'], ['02-01', '02-28'], ['03-01', '03-31'],\
                    ['04-01', '04-30'], ['05-01', '05-31'], ['06-01', '06-30'], ['07-01', '07-31'],\
                    ['08-01', '08-31'], ['09-01', '09-30'], ['10-01', '10-31'], ['11-01', '11-30'], ['12-01', '12-31']]
stock_scaler = joblib.load('./scaler/scaler1.gz')


with open(config_path, 'r') as file:
    stockinfo_config = yaml.safe_load(file)
data_api = DataLoader()
data_api.login_by_token(api_token = stockinfo_config['api_token'])
print('success')



# Flask
app = Flask(__name__)
@app.route("/")
def hello():
    return "Hello, World!"

@app.route("/StockPredict", methods=['POST'])
def searchstock():
    stocknum = request.get_json()['stocknum']
    search_month = request.get_json()['stockmonth']
    #print(stocknum, search_month)
    stockstart = '2024-' + month_start_end[search_month][0]
    stockstop = '2024-' + month_start_end[search_month][1]
    df1 = data_api.taiwan_stock_daily(
    stock_id=stocknum,
    start_date= stockstart,
    end_date=stockstop )
    if df1.shape[0] < 10:
        
        stockstart = '2024-' + month_start_end[search_month-1][0]
        stockstop = '2024-' + month_start_end[search_month-1][1]
        df2 = data_api.taiwan_stock_daily(
        stock_id=stocknum,
        start_date= stockstart,
        end_date=stockstop )
        if df1.empty:
            df1 = df2
        
        else:
            df1 = pd.concat([df2, df1],ignore_index=True)
        #print(df1)
    stock_dict = []
    for i in range(len(df1['open'])-10, len(df1['open'])):
        stockres_list = []
        for k in ['Trading_Volume', 'Trading_money', 'open', 'max', 'min', 'close', 'spread', 'Trading_turnover']:
            stockres_list.append(float(df1[k][i]))
            
        stock_dict.append(stockres_list)
    stock_trans = stock_scaler.transform(stock_dict)
    stock_tomodel = [stock_trans]
    #df1 = df1['open'][-10:]
    #df_p = pd.DataFrame({'stock_result': df1})
    #print(len(stock_dict), len(stock_dict['instances']), len(stock_dict['instances'][0]))
    data = {'instances': stock_tomodel}
    resp1 = requests.post(url = 'http://' + stockinfo_config['PR_IP'] + ':8501/v1/models/PredictionModel:predict', data = json.dumps(data, cls = NumpyEncoder))
    pr_result1 = resp1.json()['predictions'][0][0]
    to_rev = [[0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, pr_result1]]
    reved = stock_scaler.inverse_transform(to_rev)
    #print(pr_result1, reved[0][7])
    pred_price = max(0, reved[0][7])
    resp2 = requests.post(url = 'http://' + stockinfo_config['AE_IP'] + ':8501/v1/models/AutoEncoderModel:predict', data = json.dumps(data, cls = NumpyEncoder))
    ae_result2 = resp2.json()['predictions']
    recon_loss = (1.0-((np.array(stock_tomodel[0]) -np.array(ae_result2[0]))**2).mean())*100
    #print(recon_loss)
    if recon_loss < 0:
        recon_loss = 0
    return Response(json.dumps({'predictedprice': pred_price, 'predictionconfidence': recon_loss}), mimetype='application/json')


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8902)
