import { Injectable } from '@angular/core';

@Injectable()
export class DocumentService {

    // constructor(){
    //     console.log('init window service');
    // }

    get documentRef(): Document {
        return document;
    }

}
