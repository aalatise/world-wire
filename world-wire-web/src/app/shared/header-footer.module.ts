import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FlexLayoutModule } from '@angular/flex-layout';
import { RouterModule } from '@angular/router';
import {SiteHeaderComponent } from '../site-header/site-header.component';
import { PortalNavComponent } from '../portal-nav/portal-nav.component';
import { PublicNavComponent } from '../public-nav/public-nav.component';
import { PublicFooterComponent } from '../public-footer/public-footer.component';
import { CustomMaterialModule } from './custom-material.module';

/**
 * This shared module imports all default header and footer components
 *
 * @export
 * @class HeaderFooterModule
 */
@NgModule({
    imports: [
        CommonModule,
        RouterModule,
        FlexLayoutModule,
        CustomMaterialModule
    ],
    declarations: [
        SiteHeaderComponent,
        PortalNavComponent,
        PublicNavComponent,
        PublicFooterComponent
    ],
    exports: [
        SiteHeaderComponent,
        PortalNavComponent,
        PublicNavComponent,
        PublicFooterComponent
    ]
})
export class HeaderFooterModule { }
