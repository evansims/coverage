package example;

import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.*;

class MathTest {
    @Test
    void testAdd() {
        Math math = new Math();
        assertEquals(5, math.add(2, 3));
    }

    @Test
    void testIsPositive() {
        Math math = new Math();
        assertTrue(math.isPositive(1));
        assertFalse(math.isPositive(-1));
        assertFalse(math.isPositive(0));
    }
}
