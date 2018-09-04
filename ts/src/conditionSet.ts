import AuthrError from './authrError';
import Condition from './condition';
import IResource from './resource';
import { $authr, IJSONSerializable, IEvaluator, isArray, isString, empty } from './util';
import { isPlainObject } from 'lodash';

enum Conjunction {
    AND = '$and',
    OR = '$or'
}

const IMPLIED_CONJUNCTION = Conjunction.AND;

interface IConditionSetInteral {
    evaluators: IEvaluator[],
    conjunction: Conjunction
}

export default class ConditionSet implements IJSONSerializable, IEvaluator {

    private [$authr]: IConditionSetInteral = {
        conjunction: IMPLIED_CONJUNCTION,
        evaluators: []
    };

    constructor(spec: any) {
        if (isPlainObject(spec)) {
            const [conj] = Object.keys(spec);
            if (conj !== Conjunction.AND && conj !== Conjunction.OR) {
                throw new AuthrError(`Unknown condition set conjunction: ${conj}`);
            }
            this[$authr].conjunction = conj;
            spec = spec[conj];
        }
        if (!isArray(spec)) {
            throw new AuthrError('ConditionSet only takes an object or array during construction');
        }
        for (let rawe of spec) {
            if (empty(rawe)) {
                continue;
            }
            if (isArray(rawe) && rawe.length === 3 && isString(rawe[0])) {
                const [l, o, r] = rawe;
                this[$authr].evaluators.push(new Condition(l, o, r));
            } else {
                this[$authr].evaluators.push(new ConditionSet(rawe));
            }
        }
    }

    evaluate(resource: IResource): boolean {
        var result = true; // Vacuous truth: https://en.wikipedia.org/wiki/Vacuous_truth
        for (let evaluator of this[$authr].evaluators) {
            let evalResult = evaluator.evaluate(resource);
            switch (this[$authr].conjunction) {
                case Conjunction.OR:
                    if (evalResult) {
                        return true; // short circuit
                    }
                    result = false;
                    break;
                case Conjunction.AND:
                    if (!evalResult) {
                        return false; // short circuit
                    }
                    result = true;
                    break;
            }
        }
        return result;
    }

    toJSON(): any {
        var out: any = this[$authr].evaluators.map((e: any) => e.toJSON());
        if (this[$authr].conjunction !== IMPLIED_CONJUNCTION) {
            out = { [this[$authr].conjunction]: out };
        }
        return out;
    }
}
