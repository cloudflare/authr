import {
    isPlainObject as _isPlainObject,
    isObject as _isObject,
    isString as _isString,
    isArray as _isArray,
    keys as _keys,
    values as _values
} from 'lodash';
import AuthrError from './authrError';
import ISubject, { isSubject } from './subject';
import IResource, { isResource } from './resource';

export interface IEvaluator {
    evaluate(resource: IResource): boolean;
}

export interface IOperator {
    compute(left: any, right: any): boolean;
}

export interface IJSONSerializable {
    toJSON(): any;
}

export function keys(v: object): string[] {
    if (Object.keys) {
        return Object.keys(v);
    }
    return _keys(v);
}

export function values(v: object): any[] {
    if (Object.values) {
        return Object.values(v);
    }
    return _values(v);
}

export function isPlainObject(v?: any): v is object {
    return _isPlainObject(v);
}

export function isString(v?: any): v is string {
    return _isString(v);
}

export function isObject(v?: any): v is object {
    return _isObject(v);
}

export function isArray(v?: any): v is any[] {
    return _isArray(v);
}

export function empty(v?: any): boolean {
    if (v === null) {
        return true;
    }
    if (v === undefined) {
        return true;
    }
    if (isArray(v) || isString(v)) {
        return !v.length;
    }
    if (isPlainObject(v)) {
        return !keys(v).length;
    }
    return !v;
}

export function runtimeAssertIsSubject(v?: ISubject): void {
    if (!isSubject(v)) {
        throw new AuthrError('"subject" argument does not implement mandatory subject methods');
    }
}

export function runtimeAssertIsResource(v?: IResource): void {
    if (!isResource(v)) {
        throw new AuthrError('"resource" argument does not implement mandatory resource methods');
    }
}

export const $authr = Symbol('authr.admin'); // symbol to hide internal stuff
