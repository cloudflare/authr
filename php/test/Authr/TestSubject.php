<?php

namespace Cloudflare\Test\Authr;

use Cloudflare\Authr\SubjectInterface;

class TestSubject implements SubjectInterface
{
    /** @var \Cloudflare\Authr\Rule[] */
    protected $rules = null;

    /**
     * @param \Cloudflare\Authr\Rule[] $rules
     */
    public function setRules(array $rules)
    {
        $this->rules = $rules;
    }

    public function getRules()
    {
        return $this->rules;
    }
}
