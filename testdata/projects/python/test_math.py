from math_utils import add, is_positive


def test_add():
    assert add(2, 3) == 5


def test_is_positive():
    assert is_positive(1) is True
    assert is_positive(-1) is False
