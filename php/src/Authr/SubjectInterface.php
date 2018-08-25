<?php declare(strict_types=1);

namespace Cloudflare\Authr;

interface SubjectInterface
{
    /**
     * Retrieve an ordered list of rules that belong to a subject.
     *
     * @return \Cloudflare\Authr\RuleList
     */
    public function getRules(): RuleList;
}
