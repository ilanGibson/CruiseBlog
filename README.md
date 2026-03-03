# CruiseBlog
A lightweight blog application built for sharing stories, experiences, and updates during a cruise.

## Overview
CruiseBlog is a blog platform designed for passengers to post updates, share experiences, and stay connected throughout their journey at sea.

Users can:
- Create posts
- View posts from other cruisers
- Engage with shared content
- Document their trip in real-time

## Why
I went on a two-week cruise. I wanted a project I could realistically aim to build and complete during my trip.
I wanted the project to connect directly to the environemt I was in. So I built a blog application specifically as a shared digital space exclusive to passengers onboard.

The goals were:
- Ship a clearly scoped project in two weeks
- Build something themed around the cruise experience
- Focus on usability and reducing friction to entrance for users

## Features
- Create blog posts
- Complete user anonymity
- Super simple user authentication
- Chronological post feed
- Post language censorship
- Responsive design
- Admin dashboard for post update/deletion

## Tech Stack
- **Frontend:** HTML, JavaScript, CSS
- **Backend:** Golang
- **Database:** .jsonl file (everything needed to be local)
- **Authentication:** Ip hash & Simple UUID Cookie (for reduced signin/up friction)
- **Hosting:** Hosted locally on laptop (everything needed to be local) (no internet connection)

## Prerequisites
- Golang 1.25

## Run Yourself
### Clone repo
git clone https://github.com/ilanGibson/CruiseBlog.git

## Navigate into Project
cd CruiseBlog/

## Run Command
go run main.go
## Option Flags 
- [-a] creates admin page & prints one-time use path to set admin cookie
- [-p] allows user to specify port for hosting server. default port is :8090
