export function translateValueToCheckbox(val) {
    switch (val) {
        case "1":
        case "true":
        case "checked":
        case "enable":
        case "enabled":
        case "on":
        case 1:
        case true:
            val = true;
            break;

        default:
            val = false;
            break;
    }

    return val;
}
