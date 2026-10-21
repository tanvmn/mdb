#!/bin/bash

set -eu

# ==================================================================================== #
# VARIABLES
# ==================================================================================== #

# Set timezone. timedatectl list-timezones for full list.
TIMEZONE=Asia/Ho_Chi_Minh

USERNAME=mdb

read -p "Enter password for database mdb user: " DB_PASSWORD

# Force all output to be presented in en_US for the duration of this script. This avoids
# any "setting locale failed" errors while this script is running, before we have
# installed support for all locales. Do not change this setting!
export LC_ALL=en_US.UTF-8

# ==================================================================================== #
# SCRIPT LOGIC
# ==================================================================================== #

# Enable "universe" repository.
add-apt-repository --yes universe

apt update

timedatectl set-timezone ${TIMEZONE}
apt --yes install locales-all

# Add new user with sudo prvileges.
useradd --create-home --shell "/bin/bash" --groups sudo "${USERNAME}"

# Force new password for newly created user.
passwd --delete "${USERNAME}"
chage --lastday 0 "${USERNAME}"

# Copy SSH keys from root to new user.
rsync --archive --chown=${USERNAME}:${USERNAME} /root/.ssh /home/${USERNAME}

# Configure firewall to allow SSH, HTTP, HTTPS.
ufw allow 22
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable

apt --yes install fail2ban

# Install the migrate CLI tool.
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.14.1/migrate.linux-amd64.tar.gz | tar xvz
mv migrate.linux-amd64 /usr/local/bin/migrate

# Install and configure PostgreSQL.
apt --yes install postgresql

sudo -i -u postgres psql -c "create database mdb"
sudo -i -u postgres psql -d mdb -c "create extension if not exists citext"
sudo -i -u postgres psql -d mdb -c "create role mdb with login password '${DB_PASSWORD}'"

# Add a DSN to connect to db to the system-wide environment variables in /etc/environment.
echo "DB_DSN='postgres://mdb:${DB_PASSWORD}@localhost/mdb" >> /etc/environment

# Install Caddy (see https://caddyserver.com/docs/install#debian-ubuntu-raspbian).
apt install --yes debian-keyring debian-archive-keyring apt-transport-https curl
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
chmod o+r /usr/share/keyrings/caddy-stable-archive-keyring.gpg
chmod o+r /etc/apt/sources.list.d/caddy-stable.list
apt update
apt install caddy

# Upgrade all packages. Using the --force-confnew flag means that configuration
# files will be replaced if newer ones are available.
apt --yes -o Dpkg::Options::="--force-confnew" upgrade

echo "Script complete! Rebooting..."
reboot
