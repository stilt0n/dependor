// explicit type annotations
export const foo: string = 'a string';

interface MyComponentProps {
  className?: string;
  paragraphContents?: string;
  '>': 'why' | 'would' | 'you' | 'do' | 'this?'
}
// generics
export const MyComponent: React.FC<MyComponentProps> = (props) => {
  return (
    <p className={props.className}>
      {props.paragraphContents}
    </p>
  )
}

// nested generic case
export const NestedGeneric: React.FC<Pick<MyComponentProps, 'className'>> = (props) => {
  return <div className={props.className} />
}

// all the things at once
export const Brutal: React.FC</* > */Omit</* < */MyComponentProps, '>'>> = (props) => {
  return <p>rude</p>;
}

// generic with interface
export interface LinkedListNode<T> {
  key: T;
  next?: LinkedListNode<T>;
}
// generic in class
export class LinkedList<T> {
  private head: LinkedListNode<T> | null = null;
  private length: number = 0;

  constructor(initList?: T[]) {
    if (initList && initList.length > 0) {
      initList.forEach((key) => this.add(key));
    }
  }

  add(key: T) {
    if (this.head === null) {
      this.head = { key };
      this.length++;
      return;
    }
    let current = this.head;
    while (current.next) {
      current = current.next;
    }
    current.next = { key };
    this.length++;
  }

  find(key: T) {
    let current: LinkedListNode<T> | null | undefined = this.head;
    while (current) {
      if (current.key === key) {
        return current;
      }
      current = current.next;
    }
    return null;
  }
}