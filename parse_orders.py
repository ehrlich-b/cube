#!/usr/bin/env python3
import requests
from bs4 import BeautifulSoup
import re

# Fetch the page
url = "https://mzrg.com/rubik/orders.shtml"
response = requests.get(url)
soup = BeautifulSoup(response.content, 'html.parser')

# Find the table
table = soup.find('table')
if not table:
    print("No table found!")
    exit(1)

# Parse rows
rows = table.find_all('tr')
print("Order\t3x3x3 Sequence")
print("-----\t--------------")

for row in rows[1:]:  # Skip header
    cells = row.find_all('td')
    if len(cells) >= 3:  # Make sure we have enough columns
        order = cells[0].get_text(strip=True)
        sequence_3x3 = cells[2].get_text(strip=True)  # Assuming 3x3x3 is 3rd column
        
        # Only print if we have both order and sequence
        if order and sequence_3x3 and order.isdigit():
            print(f"{order}\t{sequence_3x3}")