import { hello, another } from './hello';
import { expect } from 'chai';
import 'mocha';

describe('Hello function', () => {
  
  it('should return hello world', () => {
    const result = hello();
    expect(result).to.equal('Hello World!');
  });

  it('should also return hello world', () => {
    const result = new another().hello();
    expect(result).to.equal('Hello World!');
  });

});