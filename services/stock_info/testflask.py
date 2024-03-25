from FinMind.data import DataLoader
import yaml
from flask import Flask, render_template, request, Response
import pandas as pd
import json

# Load config
config_path = '../modules/configs/stock_info.yaml'
global data_api
global stockinfo_config

month_start_end = [ ['12-01', '12-31'], ['01-01', '01-31'], ['02-01', '02-28'], ['03-01', '03-31'],\
                    ['04-01', '04-30'], ['05-01', '05-31'], ['06-01', '06-30'], ['07-01', '07-31'],\
                    ['08-01', '08-31'], ['09-01', '09-30'], ['10-01', '10-31'], ['11-01', '11-30'], ['12-01', '12-31']]


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

@app.route("/SearchStock", methods=['POST'])
def searchstock():
    stocknum = request.get_json()['stocknum']
    search_month = int(request.get_json()['stockmonth'])
    print(stocknum, search_month)
    stockstart = '2023-' + month_start_end[search_month][0]
    stockstop = '2023-' + month_start_end[search_month][1]
    df1 = data_api.taiwan_stock_daily(
    stock_id=stocknum,
    start_date= stockstart,
    end_date=stockstop )
    if df1.shape[0] < 10:
        stockstart = '2023-' + month_start_end[search_month-1][0]
        stockstop = '2023-' + month_start_end[search_month-1][1]
        df2 = data_api.taiwan_stock_daily(
        stock_id=stocknum,
        start_date= stockstart,
        end_date=stockstop )
        df1 = df2.append(df1, ignore_index = True)

    stock_dict = {}
    for i in range(len(df1['open'])-10, len(df1['open'])):
        stockres_list = []
        for k in ['open', 'close']:
            stockres_list.append(df1[k][i])
        stock_dict[df1['date'][i]] = stockres_list

    #df1 = df1['open'][-10:]
    #df_p = pd.DataFrame({'stock_result': df1})
    return Response(json.dumps(stock_dict), mimetype='application/json')
    #return "GOOD\n"

if __name__ == '__main__':
    app.debug = True
    app.run(host='0.0.0.0', port=19982)
