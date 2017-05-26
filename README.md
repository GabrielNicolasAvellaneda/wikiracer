# wikiracer
Find path between 2 wikipedia pages

## How does it work
I use priority queue to submit the tasks to workers. The priority is the depth of links value. Each worker reads from a queue processes the task and submits results back to the queue with appropriate priority.
`Mike Tyson(1)` -> `Hello World(2)`
`Mike Tyson(1)` -> `Foo(2)`
`Foo(2)` -> `Bar(3)`

The page is represented as a linked list with a link to previous page.
`Page1(nil)`, `Page2(Page1)`, `Page3(Page2)` etc.
This approach allows very fast path finding.

The pages can be crawled with wiki [API](en.wikipedia.org/w/api.php) or by parsing HTML.
## How to run
## with docker:
```
git clone https://github.com/darkonie/wikiracer
cd wikiracer
make run

a server will be started in a docker container which would be available:
http://127.0.0.1:8081/api/v1/job
use curl to submit the job and browser to get the results.
```

## API
### GET
```
/api/v1/job           returns info for all racing jobs.
/api/v1/job/{id}      returns info for one racing job.

/debug/pprof          golang profiler.
```

### POST
```
/api/v1/job               start a new job. Response will have a job ID.
/api/v1/job/{id}/cancel   cancel a job with ID.
```

### Payload
```
{
  "timeout": "20s",
  "start_page": "Mike Tyson",
  "destination_page": "Ukraine",
  "comment": "Random comment",
  "workers": 200,
  "crawl_method": "html"
}
```
 - `timeout` is used to set the job timeout. Default to 1min.
 - `crawl_method` how to crawl, using API or parse HTML. Could be `html`, `api`. Default `api`
 - `start_page`, `destionatio_page` self explanatory. Note if `crawl_method` is `html` must match the link from webpage e.g. `Mike_Tyson`. With `api` can use spaces `Mike Tyson`.
 - `comment` arbitrary comment assosiated with a job.
 - `workers` number of workers to crawl. Default `100`.

### Example
### start a new job
```
curl -i -X POST http://127.0.0.1:8081/api/v1/job -d '{"comment": "My first job", "start_page":"Mike Tyson", "destination_page": "Greek_language", "timeout": "10m", "workers": 100, "crawl_method": "html"}'
HTTP/1.1 200 OK
Content-Type: application/json
Date: Fri, 26 May 2017 22:14:23 GMT
Content-Length: 85

{"id":"ac6620d2-4260-11e7-88c3-0242ac110002","msg":"successfully started a new job"}
```

### get job status
```
curl http://127.0.0.1:8081/api/v1/job/ac6620d2-4260-11e7-88c3-0242ac110002 | jq '.'
{
  "ac6620d2-4260-11e7-88c3-0242ac110002": {
    "path": [
      "Mike Tyson",
      "Youtube",
      "Greek_language"
    ],
    "is_running": false,
    "start_link": "Mike Tyson",
    "end_link": "Greek_language",
    "status": 0,
    "comment": "My first job",
    "start_time": "2017-05-26T22:14:23.475815385Z",
    "end_time": "2017-05-26T22:14:26.854554845Z",
    "timeout": "10m0s",
    "errors": null,
    "workers": 100,
    "duration": "3.37873946s",
    "pages_visited": 555,
    "depth": 2
  }
}
```

 - `path` the result of the job. This is the path we are looking for.
 - `is_running` indicates if the job is currently running.
 - `start_link`, `end_link`, `comment`, `timeout`, `workers` same as in request.
 - `status`
   - `0` success, page was found.
   - `1` running, the job is in progress.
   - `2` cancelled, job the was cancelled because of timeout or user request.
   - `3` unchanged, the job was created but never started.
  - `pages_visited` number of pages visited.
  - `depth` the depth of crawled links.

