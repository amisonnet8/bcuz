*[日本語](history-filter_ja.md) | **English***

# Filtering by date

When there are the records of [history.jsonl](history.jsonl) (10th to 12th of October).

## Everything

```
$ bcuz -H
2026-10-12 06:52:10    0    1.2s  .  commit  $ git commit -m 'fix: handle empty reason'
2026-10-12 06:50:00    0   120ms  .  record the whole pipe  $ sh -c 'grep -c TODO *.go | sort -t: -k2 -n'
2026-10-12 06:45:00  127     0ms  .  build  $ mkae build
2026-10-12 06:41:30  130   1m05s  backend  watch the log  $ tail -f app.log
2026-10-12 06:40:03    1    2.4s  backend  check the lint findings  $ golangci-lint run
2026-10-12 06:35:12    0    8.3s  .  run the tests  $ make test
2026-10-11 09:15:30    0    4.0s  frontend  reinstall dependencies  $ npm ci
2026-10-10 18:02:44    0   35.0s  .  full tests before release  $ make test
2026-10-10 17:55:01    0   950ms  .  check the tags  $ git tag --list 'v0.*'
```

## One day only

```
$ bcuz -H --date 2026-10-10
2026-10-10 18:02:44    0   35.0s  .  full tests before release  $ make test
2026-10-10 17:55:01    0   950ms  .  check the tags  $ git tag --list 'v0.*'
```

## From a day

```
$ bcuz -H --since 2026-10-11
2026-10-12 06:52:10    0    1.2s  .  commit  $ git commit -m 'fix: handle empty reason'
2026-10-12 06:50:00    0   120ms  .  record the whole pipe  $ sh -c 'grep -c TODO *.go | sort -t: -k2 -n'
2026-10-12 06:45:00  127     0ms  .  build  $ mkae build
2026-10-12 06:41:30  130   1m05s  backend  watch the log  $ tail -f app.log
2026-10-12 06:40:03    1    2.4s  backend  check the lint findings  $ golangci-lint run
2026-10-12 06:35:12    0    8.3s  .  run the tests  $ make test
2026-10-11 09:15:30    0    4.0s  frontend  reinstall dependencies  $ npm ci
```

## A range

```
$ bcuz -H --since 2026-10-10 --until 2026-10-11
2026-10-11 09:15:30    0    4.0s  frontend  reinstall dependencies  $ npm ci
2026-10-10 18:02:44    0   35.0s  .  full tests before release  $ make test
2026-10-10 17:55:01    0   950ms  .  check the tags  $ git tag --list 'v0.*'
```

## No match

Nothing is printed, and it succeeds (0).

```
$ bcuz -H --date 2026-09-01
$ echo $?
0
```

## Mistakes

```
$ bcuz -H --date 2026-02-30
bcuz: invalid date: 2026-02-30
$ echo $?
125
$ bcuz -H --date 2026-10-12 --since 2026-10-01
bcuz: --date cannot be used with --since or --until
$ echo $?
125
```

## Time zones

The records keep `ts` in UTC. The date and time column and the comparison of dates use the time zone of **the viewer**. A record whose `ts` is `2026-10-11T21:35:12.345Z` is shown as 06:35:12 of the 12th of October in Asia/Tokyo, and as follows in a UTC environment (`TZ=UTC`), where it counts as a record of the 11th of October.

```
$ TZ=UTC bcuz -H --date 2026-10-11
2026-10-11 21:52:10    0    1.2s  .  commit  $ git commit -m 'fix: handle empty reason'
2026-10-11 21:50:00    0   120ms  .  record the whole pipe  $ sh -c 'grep -c TODO *.go | sort -t: -k2 -n'
2026-10-11 21:45:00  127     0ms  .  build  $ mkae build
2026-10-11 21:41:30  130   1m05s  backend  watch the log  $ tail -f app.log
2026-10-11 21:40:03    1    2.4s  backend  check the lint findings  $ golangci-lint run
2026-10-11 21:35:12    0    8.3s  .  run the tests  $ make test
2026-10-11 00:15:30    0    4.0s  frontend  reinstall dependencies  $ npm ci
```
