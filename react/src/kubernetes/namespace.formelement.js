import React from 'react';
import BaseComponent from '../base';

class NamespaceFormelement extends BaseComponent {
    constructor(props) {
        super(props);
        this.state = {
            _htmlid: "" + this.generateHtmlId(props.setting.name),
        };
    }

    renderInput() {
        return (
            <div className="form-group">
                <label htmlFor={this.state._htmlid} className="inputRg">{this.props.setting.label}</label>
                <input type="text" name={this.props.setting.name} id={this.state._htmlid} className="form-control"
                       placeholder={this.props.setting.placeholder} required={this.props.setting.required}
                       value={this.props.value} onChange={this.props.onchange}/>
                <div className="form-text">{this.props.setting.description}</div>
                <div className="form-text">{this.props.setting.k8sType}: {this.props.setting.k8sName}</div>
            </div>
        );
    }

    renderCheckbox() {
        let checkboxState = false;

        // translate value
        switch (this.props.value) {
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
            <div className="form-check">
                <input type="checkbox" name={this.props.setting.name} id={this.state._htmlid}
                       className="form-check-input" placeholder={this.props.setting.plaeholder}
                       required={this.props.setting.required} checked={checkboxState} onChange={this.props.onchange}/>
                <label htmlFor={this.state._htmlid} className="form-check-label">{this.props.setting.label}</label>
                <div className="form-text">{this.props.setting.description}</div>
                <div className="form-text">{this.props.setting.k8sType}: {this.props.setting.k8sName}</div>
            </div>
        );
    }

    render() {
        if (!this.props.setting.name || !this.props.setting.type) {
            return (<div></div>);
        }

        switch (this.props.setting.type) {
            case "input":
                return this.renderInput();
            case "checkbox":
                return this.renderCheckbox();
            default:
                return (<div></div>);
        }
    }
}

export default NamespaceFormelement;

