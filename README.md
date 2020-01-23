# eanclient
As simple as possible REST client to eanserver reading GTIN from serial (scanner) and retrieving product

This program opens serial port specified in the config, reads EAN/GTIN code from the scanner and sends REST
call to the eanserver. Gets Productview objects and uses simplename to add to to ShoppingList from Home Assistant
using API call as described here:
https://developers.home-assistant.io/docs/en/external_api_rest.html

This client requries server.key, to securily connect to eanserver. I'm using self-signed one copied from:
https://github.com/m4tek/eanserver

