<?php declare(strict_types=1);

namespace Cloudflare\Authr;

use Cloudflare\Authr\Condition\Operator;

final class Condition implements EvaluatorInterface
{
    const OPERATORS_CLASSES = [
        Operator\ArrayDifference::class,
        Operator\ArrayIntersect::class,
        Operator\Equals::class,
        Operator\In::class,
        Operator\Like::class,
        Operator\NotEquals::class,
        Operator\NotIn::class,
        Operator\RegExp\CaseInsensitive::class,
        Operator\RegExp\CaseSensitive::class,
        Operator\RegExp\InverseCaseInsensitive::class,
        Operator\RegExp\InverseCaseSensitive::class,
    ];

    /** @var \Cloudflare\Authr\Condition\OperatorInterface[] */
    private static $operators;

    /** @var \Cloudflare\Authr\Condition\OperatorInterface */
    private $operator;

    /** @var mixed */
    private $left;

    /** @var mixed */
    private $right;

    /**
     * @param mixed $left
     * @param string $op
     * @param mixed $right
     */
    public function __construct($left, string $op, $right)
    {
        static::initDefaultOperators();
        if (!array_key_exists($op, static::$operators)) {
            throw new Exception\InvalidConditionOperator("Unknown condition operator: '$op'");
        }
        $this->operator = static::$operators[$op];
        $this->left = $left;
        $this->right = $right;
    }

    /**
     * Evaluate a condition on a resource
     * 
     * @param \Cloudflare\Authr\ResourceInterface $resource
     * @return boolean
     */
    public function evaluate(ResourceInterface $resource): bool
    {
        return call_user_func(
            $this->operator,
            static::determineValue($resource, $this->left),
            static::determineValue($resource, $this->right)
        );
    }

    /**
     * Determine if the value passed is referring to an attribute on the resource
     * or is just the literal value.
     * 
     * @param \Cloudflare\Authr\ResourceInterface $resource
     * @param mixed $value
     * @return mixed
     */
    private static function determineValue(ResourceInterface $resource, $value)
    {
        if (is_string($value) && strlen($value) > 1) {
            if ($value[0] === '@') {
                return $resource->getResourceAttribute(substr($value, 1));
            }
            // check for escaped '@' characters and remove the escape character
            if (substr($value, 0, 2) === '\@') {
                return substr($value, 1);
            }
        }
        return $value;
    }

    protected static function initDefaultOperators()
    {
        if (is_null(static::$operators)) {
            foreach (static::OPERATORS_CLASSES as $handlerClass) {
                /** @var \Cloudflare\Authr\Condition\OperatorInterface */
                $handler = new $handlerClass;
                static::$operators[$handler->jsonSerialize()] = $handler;
            }
        }
    }

    public function jsonSerialize()
    {
        return [$this->left, $this->operator->jsonSerialize(), $this->right];
    }
}
