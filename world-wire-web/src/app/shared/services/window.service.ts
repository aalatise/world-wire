import { Injectable } from '@angular/core';

@Injectable()
export class WindowService {

    // constructor(){
    //     console.log('init window service');
    // }

    get windowRef(): Window {
        return window;
    }

}
