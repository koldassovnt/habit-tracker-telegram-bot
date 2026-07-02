# Habit tracker telegram bot for personal usage

## Tech stack
* Go as a main programming language
* PostgreSQL as a database
* Telegram bot as a client

## Features
* Personal telegram bot
* Track a habit as many times a day as you want — each tracking is counted, not just done/not-done
* A category for each habit
* Automatic text report sent every Monday and on the 1st of each month, with a per-day breakdown per habit for the week/month that just ended
* Tracked habits status check per day

## Bot interaction

### Menu Commands
* /managecategory - add, edit, delete category
* /managehabit - add, edit, delete habit
* /trackhabit - track a habit
* /todaystatus - get a status of tracked habits for today as a text
* /help - show available commands

### Menu Buttons
* Add category/habit — bot asks for a name, next message you send creates it
* Edit category/habit — pick which one from a list, then send the new name
* Delete category/habit — pick which one from a list, deleted immediately (no confirmation step). Deleting a category also removes its habits.
* Track habit — pick a category, then pick a habit, tracked instantly (can be tapped multiple times per day)

## Limitations
- Maximum 5 categories
- Maximum 10 habits in one category
- Weekly/monthly report schedule and "today" day boundary use the server's local timezone — there's no per-user timezone setting

## Running the bot

* [LOCAL_RUN.md](LOCAL_RUN.md) — run without Docker (Go + a local/dockerized Postgres)
* [DOCKER.md](DOCKER.md) — build and run the full stack with Docker Compose