<?php

namespace Cloudflare\Authr;

interface EvaluatorInterface extends \JsonSerializable
{
    /**
     * Evaluate something against a resource
     * 
     * @param \Cloudflare\Authr\ResourceInterface $resource
     * @return boolean
     */
    public function evaluate(ResourceInterface $resource);
}
