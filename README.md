This package allows you to quickly create a one-window application using [astilectron](https://github.com/asticode/go-astilectron).

Check out the [demo](https://github.com/asticode/go-astilectron-demo) to see a working example with the [bundler](https://github.com/asticode/go-astilectron-bundler).

# Installation

Run the following command:

    $ go get -u github.com/asticode/go-astilectron-bootstrap

# Prerequisites

## Files structure

You must follow the following files structure:

    |--+ resources
        |
        |--+ app (contains your static files such as .html, .css, .js, .png, etc.)
            |
            |--+ css (not mandatory, for example purposes)
            |
            |--+ html (not mandatory, for example purposes)
            |
            |--+ js (not mandatory, for example purposes)
    |--+ main.go
    
# Example

Check out the [demo](https://github.com/asticode/go-astilectron-demo) that uses the bootstrap and the [bundler](https://github.com/asticode/go-astilectron-bundler).