import React from 'react';
import ReactDOM from 'react-dom';
import App from './App';

window.$ = window.jQuery = require("jquery");
(function () {
    let rootEl = document.getElementById("root");
    if (rootEl) {
        ReactDOM.render(<App/>, rootEl);
    }

    window.$("#global-search").on('input', (event) => {
        window.App.triggerSearch(event);
        event.preventDefault();
        event.stopPropagation();
        return false;
    });
    window.$(document).on('submit', 'form', (event) => {
        event.preventDefault();
        event.stopPropagation();
        return false;
    });
})();
