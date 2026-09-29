FROM mcr.microsoft.com/playwright:v1.63.0-noble
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends dunst xdotool scrot dbus-x11 fonts-noto-cjk && rm -rf /var/lib/apt/lists/*
