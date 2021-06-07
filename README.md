# AQI traffic light
A traffic light that shows currentl AQI (air quality index)
This project is a part of a wider set of software tools to monitor air parameters and weather.

## Installation

1. Install dependencies
    ```
    sudo apt update
    sudo apt-get install python3-pip
    pip3 install RPi.GPIO
    ```
2. Run
    To run run data colloctor
    ```
    go run main.go
    ```
    Make sure that the file it generates is available over http, lighttpd can be used for this.
    ```
    python traffic-aqi.py
    ```
