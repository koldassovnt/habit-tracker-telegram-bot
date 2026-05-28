# Habit tracker telegram bot for personal usage

## Tech stack
* Go as a main programming language
* PostgreSQL as a database
* Telegram bot as a client

## Features
* Personal telegram bot
* Track habit as DONE, UNDONE + How many times it was done per day
* A category for each habit
* Tracked habits report as text or files: xlsx, pdf
* Tracked habits status check

## Bot interaction

### Commands
* /managecategory - add, edit, delete category
* /managehabit - add, edit, delete habit
* /trackhabit - track a habit
* /todaystatus - get a status of tracket habit for today as a text
* /report - get a report text or file of some date range

### Menu Buttons
...

## Limitations
- Maximum 5 categories
- Maximum 10 habits in one category
- Report range options: day, current week, current month, choose a month

## Run in personal computer using docker as a telegram bot server

### Step 1 Create a telegram bot using @BotFather
...

### Step 2 Set up project and create a docker image
...

### Step 3 Set up docker container and run it
...