<?php

namespace Example\Tests;

use Example\Math;
use PHPUnit\Framework\TestCase;

class MathTest extends TestCase
{
    public function testAdd(): void
    {
        $math = new Math();
        $this->assertSame(5, $math->add(2, 3));
    }

    public function testIsPositive(): void
    {
        $math = new Math();
        $this->assertTrue($math->isPositive(1));
        $this->assertFalse($math->isPositive(-1));
    }
}
