import yfinance as yf
from datetime import datetime, timedelta
import os
import csv

# CHANGE THIS TO RIGHT TICKER
ticker = "TMF"


# Get today's date
end_date = datetime.now()

start_date = end_date - timedelta(days=365 * 150)


start_data = start_date.strftime('%Y-%m-%d')
end_date = end_date.strftime('%Y-%m-%d')

output_dir = os.getcwd() 

data = yf.download(ticker, start=start_date, end=end_date)
os.makedirs(output_dir, exist_ok=True)  # Create the directory if it doesn't exist
csv_path = os.path.join(output_dir, f'{ticker}.csv')
# Export data to a CSV file
data.to_csv(csv_path)
