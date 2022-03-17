import React from 'react';
import ReactDOM from 'react-dom';
import App from './app';

window.$ = window.jQuery = require("jquery");
(function () {
    /**
     * Object.prototype.forEach() polyfill
     * https://gomakethings.com/looping-through-objects-with-es6/
     * @author Chris Ferdinandi
     * @license MIT
     */
    if (!Object.prototype.forEach) {
        Object.defineProperty(Object.prototype, 'forEach', {
            value: function (callback, thisArg) {
                if (this == null) {
                    throw new TypeError('Not an object');
                }
                thisArg = thisArg || window;
                for (var key in this) {
                    if (this.hasOwnProperty(key)) {
                        callback.call(thisArg, this[key], key, this);
                    }
                }
            }
        });
    }

    // init global settings
    window.APP_VERSION = window.$("html").attr("app:version");
    window.CSRF_TOKEN = window.$("html").attr("app:csrf");
    window.$("html").attr("app:csrf", "");

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
