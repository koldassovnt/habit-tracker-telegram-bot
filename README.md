# Habit tracker telegram bot for personal usage

## Tech stack
* Go as a main programming language
* PostgreSQL as a database
* Telegram bot as a client

## Features
* Personal telegram bot
* Track habit as DONE and How many times it was done per day
* A category for each habit
* Send tracked habits report as text and pdf file once a week and once a month
* Tracked habits status check per day

## Bot interaction

### Menu Commands
* /managecategory - add, edit, delete category
* /managehabit - add, edit, delete habit
* /trackhabit - track a habit
* /todaystatus - get a status of tracked habits for today as a text

### Menu Buttons
...

## Limitations
- Maximum 5 categories
- Maximum 10 habits in one category

## Run in personal computer using docker as a telegram bot server

### Step 1 Create a telegram bot using @BotFather
...

### Step 2 Set up project and create a docker image
...

### Step 3 Set up docker container and run it
...

### Run for local testing
export $(cat .env | xargs) && go run main.go