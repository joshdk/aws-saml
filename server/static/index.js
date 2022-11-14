"use strict";

// ready calls the given function when the page is loaded and ready.
function ready(fn) {
    if (document.readyState !== 'loading') {
        fn();
    } else {
        document.addEventListener('DOMContentLoaded', fn);
    }
}

// shutdown sends a POST to the shutdown endpoint, signaling to the server to
// terminate itself.
function shutdown() {
    fetch('/shutdown', {method: 'POST'});
}

ready(shutdown);
