<?php

namespace Cloudflare\Authr;

interface SubjectInterface
{
    /**
     * Retrieve an ordered list of rules that belong to a subject.
     *
     * @return \Cloudflare\Authr\Rule[]
     */
    public function getRules();
}
