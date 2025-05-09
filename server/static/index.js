"use strict";

window.addEventListener('load', () => {
    fetch('/response')
    .then(response => response.text()) // send response body to next then chain
    .then(body => {
        console.groupCollapsed("SAML Assertion");
        console.log(body);
        console.groupEnd();
    })
    .finally(() => {
        fetch('/shutdown', {method: 'POST'});
    });
});
