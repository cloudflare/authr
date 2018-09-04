import {
    $authr,
    IJSONSerializable,
    IEvaluator,
    isString,
    isArray
} from './util';
import { findIndex, intersectionWith } from 'lodash';
import IResource, { SYM_GET_RSRC_ATTR } from './resource';
import AuthrError from './authrError';

interface IOperatorFunc {
    (left: any, right: any): boolean
}

export enum OperatorSign {
    EQUALS = '=',
    NOT_EQUALS = '!=',
    LIKE = '~=',
    CASE_SENSITIVE_REGEXP = '~',
    CASE_INSENSITIVE_REGEXP = '~*',
    INV_CASE_SENSITIVE_REGEXP = '!~',
    INV_CASE_INSENSITIVE_REGEXP = '!~*',
    IN = '$in',
    NOT_IN = '$nin',
    ARRAY_INTERSECT = '&',
    ARRAY_DIFFERENCE = '-'
}

const operators: Map<OperatorSign, IOperatorFunc> = new Map([
    [
        OperatorSign.EQUALS,
        (left: any, right: any): boolean => left == right
    ],
    [
        OperatorSign.NOT_EQUALS,
        (left: any, right: any): boolean => left != right
    ],
    [
        OperatorSign.LIKE,
        (left: any, right: any): boolean => {
            let pleft = '^';
            let pright = '$';
            right = `${right}`;
            if (right.startsWith('*')) {
                pleft = '';
            }
            if (right.endsWith('*')) {
                pright = '';
            }
            return RegExp(`${pleft}${right.replace('*', '')}${pright}`, 'i').test(left);
        }
    ],
    [
        OperatorSign.CASE_SENSITIVE_REGEXP,
        (left: any, right: any): boolean => RegExp(right).test(left)
    ],
    [
        OperatorSign.CASE_INSENSITIVE_REGEXP,
        (left: any, right: any): boolean => RegExp(right, 'i').test(left)
    ],
    [
        OperatorSign.INV_CASE_SENSITIVE_REGEXP,
        (left: any, right: any): boolean => !RegExp(right).test(left)
    ],
    [
        OperatorSign.INV_CASE_INSENSITIVE_REGEXP,
        (left: any, right: any): boolean => !RegExp(right, 'i').test(left)
    ],
    [
        OperatorSign.IN,
        (left: any, right: any): boolean => {
            if (!isArray(right)) {
                return false;
            }
            return findIndex(right, (v: any) => v == left) >= 0;
        }
    ],
    [
        OperatorSign.NOT_IN,
        (left: any, right: any): boolean => {
            if (!isArray(right)) {
                return false;
            }
            return findIndex(right, (v: any) => v == left) === -1;
        }
    ],
    [
        OperatorSign.ARRAY_INTERSECT,
        (left: any, right: any): boolean => {
            if (!isArray(left) || !isArray(right)) {
                return false;
            }
            return intersectionWith(left, right, (a: any, b: any) => a == b).length > 0;
        }
    ],
    [
        OperatorSign.ARRAY_DIFFERENCE,
        (left: any, right: any): boolean => {
            if (!isArray(left) || !isArray(right)) {
                return false;
            }
            return intersectionWith(left, right, (a: any, b: any) => a == b).length === 0;
        }
    ]
]);

function determineValue(resource: IResource, value: any): any {
    if (isString(value) && value.length > 1) {
        if (value.charAt(0) === '@') {
            return resource[SYM_GET_RSRC_ATTR](value.substr(1));
        }
        if (value.substr(0, 2) === '\\@') {
            return value.substr(1);
        }
    }
    return value;
}

interface IConditionInternal {
    left: any;
    right: any;
    sign: OperatorSign,
    operator: IOperatorFunc
}

export default class Condition implements IJSONSerializable, IEvaluator {

    private [$authr]: IConditionInternal;

    constructor(left: any, opsign: string, right: any) {
        const op = operators.get(opsign as OperatorSign);
        if (!op) {
            throw new AuthrError(`Unknown condition operator: '${opsign}'`);
        }
        this[$authr] = {
            left, right,
            sign: opsign as OperatorSign,
            operator: op
        };
    }

    evaluate(resource: IResource): boolean {
        return this[$authr].operator(
            determineValue(resource, this[$authr].left),
            determineValue(resource, this[$authr].right)
        );
    }

    toJSON(): any {
        const { left, sign, right } = this[$authr];
        return [left, sign, right];
    }
}
