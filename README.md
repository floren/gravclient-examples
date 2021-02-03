# Gravwell Client Library examples

This respository contains examples demonstrating the use of the [Gravwell client library](https://pkg.go.dev/github.com/gravwell/gravwell/v3/client).

* runsearch: Runs a search query using the text renderer and displays the results.
* barchart: Runs a search query with the chart renderer and outputs commands which can be fed to gnuplot to generate a bar chart.
* backup: Generates a backup of all content on a Gravwell webserver.
* restore: Restores a webserver using the output of the 'backup' command.
