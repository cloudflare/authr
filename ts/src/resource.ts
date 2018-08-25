import { isObject } from 'lodash';

export const SYM_GET_RSRC_TYPE = Symbol('authr.resource_get_resource_type');
export const SYM_GET_RSRC_ATTR = Symbol('authr.resource_get_resource_attribute');

interface IResource {
    [SYM_GET_RSRC_TYPE](): string;
    [SYM_GET_RSRC_ATTR](attribute: string): any;
}

export function isResource(v?: any): v is IResource {
    if (!isObject(v)) {
        return false;
    }
    return v.hasOwnProperty(SYM_GET_RSRC_ATTR) &&
        v.hasOwnProperty(SYM_GET_RSRC_TYPE);
}

export default IResource;
