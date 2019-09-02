# Errors

The Pool Detective API uses the following error codes:

Error Code | Meaning
---------- | -------
400 | Bad Request -- Your request is invalid.
401 | Unauthorized -- Your API key is wrong.
403 | Forbidden -- The resource you're trying to open is not accessible to your API key
404 | Not Found -- The resource you're trying to open is not found
405 | Method Not Allowed -- You tried to access a resource using the wrong HTTP verb
406 | Not Acceptable -- You requested a format that isn't json.
429 | Too Many Requests -- You've hit the rate limit
500 | Internal Server Error -- We had a problem with our server. Try again later.
503 | Service Unavailable -- We're temporarily offline for maintenance. Please try again later.
