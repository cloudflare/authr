<?php declare(strict_types=1);

namespace Cloudflare\Authr;

use JsonSerializable;

final class ConditionSet implements EvaluatorInterface
{
    const LOGICAL_AND = '$and';
    const LOGICAL_OR = '$or';

    /**
     * If ConditionSet receives an array that is not prepended with a
     * conjunction, it will infer the logical conjunction.
     */
    const IMPLIED_CONJUNCTION = self::LOGICAL_AND;

    /** @var \Cloudflare\Authr\EvaluatorInterface[] */
    private $evaluators = [];

    /** @var string */
    private $conjunction = self::IMPLIED_CONJUNCTION;

    /**
     * @param array $spec
     */
    public function __construct($spec)
    {
        if (!is_array($spec)) {
            throw new Exception\InvalidConditionSetException('ConditionSet only takes an array during construction');
        }
        if (in_array(key($spec), [static::LOGICAL_OR, static::LOGICAL_AND], true)) {
            $this->conjunction = key($spec);
            $spec = $spec[key($spec)];
        }
        foreach ($spec as $rawEvaluator) {
            if (empty($rawEvaluator) || !is_array($rawEvaluator)) {
                continue;
            }
            if (count($rawEvaluator) === 3 && is_string($rawEvaluator[1])) {
                // this is probably a condition
                list($attr, $op, $val) = array_values($rawEvaluator);
                $this->evaluators[] = new Condition($attr, $op, $val);
                continue;
            }
            // probably a nested condition set, let a recursive construction do
            // more validation
            $this->evaluators[] = new static($rawEvaluator);
        }
    }

    /**
     * {@inheritDoc}
     */
    public function evaluate(ResourceInterface $resource): bool
    {
        $result = true; // Vacuous truth: https://en.wikipedia.org/wiki/Vacuous_truth
        foreach ($this->evaluators as $evaluator) {
            $evalResult = $evaluator->evaluate($resource);
            if (!is_bool($evalResult)) {
                $t = gettype($evalResult);
                throw new Exception\RuntimeException("Unexpected value encountered while evaluating conditions. Expected boolean, received $t");
            }
            if ($this->conjunction === static::LOGICAL_OR) {
                if ($evalResult) {
                    return true; // short circuit
                }
                $result = false;
            }
            if ($this->conjunction === static::LOGICAL_AND) {
                if (!$evalResult) {
                    return false; // short circuit
                }
                $result = true;
            }
        }

        return $result;
    }

    public function jsonSerialize()
    {
        $result = [];
        foreach ($this->evaluators as $evaluator) {
            $result[] = $evaluator->jsonSerialize();
        }
        if ($this->conjunction !== static::IMPLIED_CONJUNCTION) {
            $result = [$this->conjunction => $result];
        }

        return $result;
    }
}
