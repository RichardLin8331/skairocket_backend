FROM python:3.8-slim

WORKDIR /scaler
COPY ./scaler/scaler1.gz .

WORKDIR /configs
COPY ./configs/stock_info.yaml .

WORKDIR /

COPY ./requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

COPY ./stock_info.py ./

CMD [ "python", "stock_info.py" ]