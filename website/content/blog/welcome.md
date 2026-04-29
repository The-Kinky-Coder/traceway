---
title: "Welcome to the Traceway Engineering Blog"
date: 2026-04-29
author: "The Traceway Team"
excerpt: "We're starting a blog. Release notes, deep-dives into the platform, and the engineering decisions behind Traceway."
tags: ["announcement"]
---

We're starting a blog.

This is where we'll share **release notes**, deep-dives into how Traceway works under the hood, and the design choices that shape the product. If you're running observability for your team, we hope it's useful.

## What you'll find here

- **Release notes** — every meaningful change, what it does, and why it matters.
- **Engineering deep-dives** — the architecture choices behind logs, traces, metrics, replay, and exceptions.
- **Performance work** — how we keep ingestion fast and queries cheap on ClickHouse.

## A quick example

Here's the kind of code you'll see in posts about the SDKs:

```go
import "github.com/tracewayapp/traceway"

func main() {
    traceway.Init("myapp", "<token>@https://cloud.tracewayapp.com/api/report")
    defer traceway.Close()

    // your app here
}
```

That's it for the welcome post. New posts will land here as we ship.
