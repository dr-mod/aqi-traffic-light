import urllib.request
import json
from gpiozero import LED
from gpiozero import Button
from time import sleep
from signal import pause
import datetime


button = Button(22)

add_green = LED(17)
add_red = LED(27)
red = LED(18)
green = LED(23)
yellow = LED(24)


def display_status(aqi_variable):
    if aqi_variable < 51:
        green.on()
        yellow.off()
        red.off()
    elif 51 <= aqi_variable <= 100:
        yellow.on()
        green.off()
        red.off()
    elif aqi_variable > 100:
        red.on()
        green.off()
        yellow.off()


def clear_display():
    red.off()
    green.off()
    yellow.off();


def fetch_show_aqi():
    external_data = json.loads(urllib.request.urlopen("http://localhost:81/aqi.json").read())
    aqi = external_data["aqi"]
    display_status(aqi)


button.when_pressed = fetch_show_aqi

while True:
    try:
        now = datetime.datetime.now()
        if now.hour >= 8:
            fetch_show_aqi()
        else:
            clear_display()
    except Exception:
        clear_display()
        pass
    sleep(60)


pause()



