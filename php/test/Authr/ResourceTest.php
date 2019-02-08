<?php

namespace Cloudflare\Test\Authr;

use Cloudflare\Test\TestCase;
use Cloudflare\Authr\Resource;
use Cloudflare\Authr\Exception\InvalidAdHocResourceException;

class ResourceTest extends TestCase
{
    public function testUnspecifiedType()
    {
        $this->expectException(InvalidAdHocResourceException::class);
        $rsrc = Resource::adhoc('', []);
    }

    public function testGetResourceType()
    {
        $rsrc = Resource::adhoc('thing', [
            'foo' => 'bar'
        ]);
        $this->assertEquals('thing', $rsrc->getResourceType());
    }

    public function testUnknownAttribute()
    {
        $rsrc = Resource::adhoc('thing', [
            'foo' => 'bar'
        ]);
        $this->assertNull($rsrc->getResourceAttribute('lol'));
    }

    public function testCallableAttribute()
    {
        $rsrc = Resource::adhoc('thing', [
            'foo' => 'bar',
            'id' => function () {
                return 198;
            }
        ]);
        $this->assertEquals('bar', $rsrc->getResourceAttribute('foo'));
        $this->assertEquals(198, $rsrc->getResourceAttribute('id'));
    }
}
