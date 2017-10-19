<?php

namespace Cloudflare;

use Cloudflare\Authr\Exception;
use Cloudflare\Authr\ResourceInterface;
use Cloudflare\Authr\Rule;
use Cloudflare\Authr\SubjectInterface;
use Cloudflare\Authr\Condition;
use Psr\Log\LoggerInterface;

/**
 * @suppress PhanUnreferencedClass
 */
final class Authr implements AuthrInterface
{
    /** @var LoggerInterface */
    private $logger;

    /**
     * A runtime cache of valid operators for rules
     * 
     * @var string[]
     */
    private static $validOperators;

    public function __construct(LoggerInterface $logger)
    {
        $this->logger = $logger;
    }

    /**
     * {@inheritDoc}
     */
    public function can(SubjectInterface $subject, $action, ResourceInterface $resource)
    {
        $rules = $subject->getRules();
        if (!is_array($rules)) {
            throw new Exception\RuntimeException('Unexpected type returned from subject rule retrieval');
        }
        $rt = $resource->getResourceType();
        $this->logger->info('checking permissions', ['action' => $action, 'rsrc_type' => $rt]);
        $i = 0;
        foreach ($rules as $rule) {
            if (!$rule instanceof Rule) {
                throw new Exception\RuntimeException('Unexpected type found in subject permission list');
            }
            if (!$rule->resourceTypes()->contains($rt)) {
                $this->logger->debug('continuing permission check, rsrc_type mismatch', ['rule_no' => ++$i]);
                continue;
            }
            if (!$rule->actions()->contains($action)) {
                $this->logger->debug('continuing permission check, action mismatch', ['rule_no' => ++$i]);
                continue;
            }
            if (!$rule->conditions()->evaluate($resource)) {
                $this->logger->debug('continuing permission check, rsrc_match mismatch', ['rule_no' => ++$i]);
                continue;
            }

            if ($rule->access() === Rule::ALLOW) {
                $this->logger->info('rule matched! allowing action...', ['rule_no' => ++$i]);
                return true;
            } else if ($rule->access() === Rule::DENY) {
                $this->logger->info('rule matched! denying action...', ['rule_no' => ++$i]);
                return false;
            }

            // unknown type!
            throw new Exception\RuntimeException(sprintf('Rule access set to unknown value: %s', strval($rule->access())));
        }

        $this->logger->info('no rules matched. denying action...', ['action' => $action, 'rsrc_type' => $rt]);
        // default to "deny all"
        return false;
    }

    /**
     * {@inheritDoc}
     */
    public function validateRule($definition)
    {
        if (!static::isMap($definition)) {
            throw new Exception\ValidationException('Rule definition must be a map');
        }
        if (array_key_exists('access', $definition)) {
            if (!in_array($definition['access'], [Rule::ALLOW, Rule::DENY], true)) {
                throw new Exception\ValidationException("Invalid access type: '{$definition['access']}'");
            }
            if (array_key_exists('where', $definition)) {
                $where = $definition['where'];
            } else {
                $where = null;
            }
        } else {
            throw new Exception\ValidationException('Rule must specify an access type');
        }
        if (!static::isMap($where)) {
            throw new Exception\ValidationException('Rule where clause must be a map');
        }
        $needWhereKeys = ['rsrc_type', 'rsrc_match', 'action'];
        $haveWhereKeys = array_keys($where);
        $diff = array_diff($needWhereKeys, $haveWhereKeys);
        if (!empty($diff)) {
            $needKeys = implode("', '", $diff);
            throw new Exception\ValidationException("Missing key(s) '$needKeys' in rule where clause");
        }
        $diff = array_diff($haveWhereKeys, $needWhereKeys);
        if (!empty($diff)) {
            $unknownKeys = implode("', '", $diff);
            throw new Exception\ValidationException("Unknown key(s) '$unknownKeys' in rule where clause");
        }
        $this->validateRuleSlugSet($where, 'action');
        $this->validateRuleSlugSet($where, 'rsrc_type');
        $this->validateRuleConditionSet($where['rsrc_match']);
    }

    /**
     * @param mixed[] $where
     * @param string $ssKey
     * @return void
     */
    private function validateRuleSlugSet($where, $ssKey)
    {
        $ss = $where[$ssKey];
        if (static::isMap($ss)) {
            $needssKeys = ['$not'];
            $havessKeys = array_keys($ss);
            $diff = array_diff($needssKeys, $havessKeys);
            if (!empty($diff)) {
                $needKeys = implode("', '", $diff);
                throw new Exception\ValidationException("Missing key '\$not' in '$ssKey' section of rule where clause");
            }
            $diff = array_diff($havessKeys, $needssKeys);
            if (!empty($diff)) {
                $unknownKeys = implode("', '", $diff);
                throw new Exception\ValidationException("Unknown key(s) '$unknownKeys' in '$ssKey' section of rule where clause");
            }
            $ss = $ss['$not'];
        }
        if (static::isList($ss)) {
            foreach ($ss as $value) {
                if (!is_string($value)) {
                    $uexptype = gettype($value);
                    throw new Exception\ValidationException("Unexpected value type '$uexptype' found in '$ssKey' section of rule where clause");
                }
            }
        } elseif (!is_string($ss)) {
            $uexptype = gettype($ss);
            throw new Exception\ValidationException("Unexpected value type '$uexptype' found in '$ssKey' section of rule where clause");
        }
    }

    /**
     * @param mixed[] $conditions
     * @return void
     */
    private function validateRuleConditionSet($conditions)
    {
        if (static::isMap($conditions)) {
            $haveCondKeys = array_keys($conditions);
            if (count($haveCondKeys) > 1) {
                $otherKeys = implode("', '", array_values(array_filter($haveCondKeys, function ($key) { return $key !== '$or' && $key !== '$and'; })));
                throw new Exception\ValidationException("Unknown key(s) '$otherKeys' found in a condition set in the 'rsrc_match' section of the rule where clause");
            } else if (count($haveCondKeys) === 0) {
                throw new Exception\ValidationException("Empty map found in a set of conditions in the 'rsrc_match' section of the rule where clause");
            }
            $logic = $haveCondKeys[0];
            if ($logic !== '$and' && $logic !== '$or') {
                throw new Exception\ValidationException("Resource conditions (rsrc_match) must have a single key ('\$and' OR '\$or', got '$logic') if it is a map");
            }
            $conditions = $conditions[$logic];
        }
        if (!static::isList($conditions)) {
            throw new Exception\ValidationException('Resource conditions (rsrc_match) is invalid');
        }
        foreach ($conditions as $idx => $value) {
            if (static::isList($value) && count($value) === 3 && is_string($value[1])) {
                // this is a single condition, just check the operator
                if (!in_array($value[1], static::getValidOperators(), true)) {
                    throw new Exception\ValidationException("Unknown operater found in a condition in 'rsrc_match': '{$value[1]}'");
                }
            } else {
                $this->validateRuleConditionSet($value);
            }
        }
    }

    /**
     * Retrieve a list of valid condition operators
     * 
     * @return string[]
     */
    private static function getValidOperators()
    {
        if (is_null(static::$validOperators)) {
            static::$validOperators = array_map(function ($cc) { return (new $cc)->jsonSerialize(); }, Condition::OPERATORS_CLASSES);
        }
        return static::$validOperators;
    }

    /** @internal */
    const EMPTY_IS_ASSOCIATIVE = 0;

    /** @internal */
    const EMPTY_IS_NOT_ASSOCIATIVE = 1;

    /**
     * isMap will inspect an array and determine if it is an associative array
     * with non-numeric keys. Optionally set how isMap should interpret empty
     * arrays with $mode.
     * 
     * @param array $arr
     * @param int $mode
     * @return boolean
     */
    private static function isMap($arr, $mode = self::EMPTY_IS_ASSOCIATIVE)
    {
        if (!is_array($arr)) {
            return false;
        }
        if (count($arr) === 0) {
            if ($mode === static::EMPTY_IS_ASSOCIATIVE) {
                return true;
            } else if ($mode === static::EMPTY_IS_NOT_ASSOCIATIVE) {
                return false;
            }
            throw new \InvalidArgumentException('invalid mode arg in isMap');
        }
        $i = 0;
        foreach ($arr as $key => $value) {
            if ($key !== $i) {
                return true;
            }
            ++$i;
        }

        return false;
    }

    /**
     * isList will inspect an array and determine if it is a list array with only
     * ordered, integer, numeric keys. Optionally set how isList should interpret
     * empty arrays with $mode.
     * 
     * @param array $arr
     * @param int $mode
     * @return boolean
     */
    private static function isList($arr, $mode = self::EMPTY_IS_NOT_ASSOCIATIVE)
    {
        if (!is_array($arr)) {
            return false;
        }
        return !static::isMap($arr, $mode);
    }
}
