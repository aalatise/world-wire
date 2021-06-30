import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { SecureVersionComponent } from './secure-version.component';

const docsRoutes: Routes = [
    {
        path: '',
        component: SecureVersionComponent
    }
];

@NgModule({
    imports: [
        RouterModule.forChild(docsRoutes)
    ],
    exports: [
        RouterModule
    ]
})
export class DocsRoutingModule { }
