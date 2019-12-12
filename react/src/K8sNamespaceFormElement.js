import React from 'react';
import BaseComponent from './BaseComponent';

class K8sNamespaceFormElement extends BaseComponent {
    constructor(props) {
        super(props);

        let htmlId = this.props.setting.Name + Math.random().toString(36).substr(2, 9);
        htmlId = htmlId.replace(/[^a-zA-Z0-9]/g, '');

        this.state = {
            htmlId: htmlId,
        };
    }

    generateHtmlId(setting) {
        return this.state.htmlId;
    }

    renderInput() {
        return (
            <div className="form-group">
                <label htmlFor={this.generateHtmlId()} className="inputRg">{this.props.setting.Label}</label>
                <input type="text" name={this.props.setting.Name} id={this.generateHtmlId()} className="form-control" placeholder={this.props.setting.Plaeholder} required={this.props.setting.Validation.Required} value={this.props.value} onChange={this.props.onchange} />
                <small className="form-text text-muted">{this.props.setting.Description}</small>
                <small className="form-text text-muted">{this.props.setting.K8sType}: {this.props.setting.K8sName}</small>
            </div>
        );
    }

    renderCheckbox() {
        let checkboxState = false;

        // translate value
        switch(this.props.value) {
            case "1":
            case "true":
            case "checked":
            case "enable":
            case "enabled":
            case "on":
            case 1:
            case true:
                checkboxState = true;
                break;
        }

        return (
            <div className="form-group form-check">
                <input type="checkbox" name={this.props.setting.Name} id={this.generateHtmlId()} className="form-check-input" placeholder={this.props.setting.Plaeholder} required={this.props.setting.Validation.Required} checked={checkboxState} onChange={this.props.onchange} />
                <label htmlFor={this.generateHtmlId()} className="form-check-label">{this.props.setting.Label}</label>
                <small className="form-text text-muted">{this.props.setting.Description}</small>
                <small className="form-text text-muted">{this.props.setting.K8sType}: {this.props.setting.K8sName}</small>
            </div>
        );
    }

    render() {
        if (!this.props.setting.Name || !this.props.setting.Type) {
            return (<div></div>);
        }

        switch (this.props.setting.Type) {
            case "input":
                return this.renderInput();
            case "checkbox":
                return this.renderCheckbox();
            default:
                return (<div></div>);
        }
    }
}

export default K8sNamespaceFormElement;

